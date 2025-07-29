package com.uber.scip.aggregator.scip;

import com.sourcegraph.Scip;
import com.sourcegraph.scip_semanticdb.ScipSemanticdbOptions;
import com.sourcegraph.scip_semanticdb.ScipWriter;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;
import java.util.stream.IntStream;
import javax.tools.Diagnostic;
import javax.tools.DiagnosticCollector;
import javax.tools.JavaFileObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * UberScipWriter is a custom implementation of the ScipWriter class that handles the generation of
 * SCIP (Source Code Information Protocol) files for Java projects.
 *
 * <p>This class extends the ScipWriter class and overrides the emitTyped method to customize the
 * output of SCIP documents. It filters out certain documents based on their relative paths and
 * modifies the document list to ensure stable output.
 */
public class UberScipWriter extends ScipWriter {

  private static final Logger logger = LoggerFactory.getLogger(UberScipWriter.class);

  private static final String SCIP_LOCAL_DIAGNOSTIC_PREFIX = "local diagnostic_";

  BuildOptions buildOptions;
  Map<String, List<CompilationIssue>> analysisIssues = Map.of();

  public UberScipWriter(
      ScipSemanticdbOptions options,
      BuildOptions buildOptions,
      DiagnosticCollector<JavaFileObject> diagnosticCollector)
      throws IOException {
    super(options);
    this.buildOptions = buildOptions;
    if (diagnosticCollector != null && !diagnosticCollector.getDiagnostics().isEmpty()) {
      this.analysisIssues =
          diagnosticCollector.getDiagnostics().stream()
              .map(CompilationIssue::new)
              //
              .collect(
                  Collectors.toMap(
                      CompilationIssue::getSource,
                      List::of,
                      (a, b) -> {
                        List<CompilationIssue> list = new ArrayList<>(a);
                        list.addAll(b);
                        return list;
                      }));
    }
  }

  private static Scip.Severity mapCompilationKindToScipSeverity(Diagnostic.Kind kind) {
    switch (kind) {
      case ERROR:
        return Scip.Severity.Error;
      case WARNING:
      case MANDATORY_WARNING:
        return Scip.Severity.Warning;
      case NOTE:
        return Scip.Severity.Information;
      case OTHER:
      default:
        return Scip.Severity.UNRECOGNIZED;
    }
  }

  @Override
  public void emitTyped(Scip.Index index) {
    // We need stable output, documents should be sorted by their relative path.
    // In java proto list is unmodifiable, so we need to create a new list.
    List<Scip.Document> processedDocuments = new ArrayList<>();
    for (Scip.Document document : index.getDocumentsList()) {
      if (isGeneratedDocument(document)) {
        if (is3rdPartyDocument(document) && !this.buildOptions.shouldShade3rdPartySymbols()) {
          processedDocuments.add(document);
          continue;
        }
        if (isIdlDocument(document) && !this.buildOptions.shouldShadeIdlSymbols()) {
          processedDocuments.add(document);
          continue;
        }
        // We need to set the relative path for each document to be consistent with the project
        // root.
        List<Scip.Occurrence> occurrences = document.getOccurrencesList();
        occurrences =
            occurrences.stream()
                .filter(
                    occurrence ->
                        isDefinition(occurrence) && !isLocalSymbol(occurrence.getSymbol()))
                .collect(Collectors.toList());

        List<Scip.SymbolInformation> symbolInformation = document.getSymbolsList();
        symbolInformation =
            symbolInformation.stream()
                .filter(symbolInfo -> !isLocalSymbol(symbolInfo.getSymbol()))
                .collect(Collectors.toList());

        processedDocuments.add(
            document.toBuilder()
                .clearOccurrences()
                .clearSymbols()
                .addAllOccurrences(occurrences)
                .addAllSymbols(symbolInformation)
                .build());
      } else {
        String path = document.getRelativePath();
        List<CompilationIssue> fileIssues = analysisIssues.get(path);
        if (this.analysisIssues.containsKey(path) && fileIssues != null) {
          Scip.Document.Builder documentBuilder = document.toBuilder();
          // No parallel execution to maintain order
          IntStream.range(0, fileIssues.size())
              .forEach(
                  i -> {
                    CompilationIssue issue = fileIssues.get(i);
                    Scip.Occurrence.Builder newOccurrence =
                        Scip.Occurrence.newBuilder()
                            .setSymbol(SCIP_LOCAL_DIAGNOSTIC_PREFIX + i)
                            .addRange((int) issue.getLineNumber()) // startLine
                            .addRange((int) issue.getColumnNumberStart()) // startCharacter
                            .addRange((int) issue.getColumnNumberEnd()) // endCharacter (estimate)
                            .addDiagnostics(
                                Scip.Diagnostic.newBuilder()
                                    .setSeverity(mapCompilationKindToScipSeverity(issue.getKind()))
                                    .setMessage(issue.getMessage())
                                    .build());

                    documentBuilder.addOccurrences(newOccurrence);
                  });
          document = documentBuilder.build();
        }
        processedDocuments.add(document);
      }
    }

    processedDocuments =
        processedDocuments.stream()
            .sorted(Comparator.comparing(Scip.Document::getRelativePath))
            .collect(Collectors.toList());
    // We are not using the default metadata builder because we want to set the project root.
    Scip.Metadata metadata = index.getMetadata();
    index =
        index.toBuilder()
            .setMetadata(metadata.toBuilder().setProjectRoot("java-code").build())
            .clearDocuments()
            .addAllDocuments(processedDocuments)
            .build();
    this.originalEmitTyped(index);
  }

  public void originalEmitTyped(Scip.Index index) {
    super.emitTyped(index);
  }

  private boolean isGeneratedDocument(Scip.Document document) {
    return document.getRelativePath().startsWith("bazel-out/")
        || document.getRelativePath().startsWith("bazel-bin/");
  }

  public static boolean isIdlDocument(Scip.Document document) {
    return document.getRelativePath().matches("^bazel-out/[^/]+/bin/idl/.*");
  }

  public static boolean is3rdPartyDocument(Scip.Document document) {
    return is3rdPartyOutputPath(document.getRelativePath());
  }

  public static boolean is3rdPartyOutputPath(String path) {
    return path.matches("^bazel-out/[^/]+/bin/3rdparty/.*");
  }

  public static boolean isLocalSymbol(String symbol) {
    return symbol.startsWith("local");
  }

  public static boolean isDefinition(Scip.Occurrence occurrence) {
    return occurrence.getSymbolRoles() == Scip.SymbolRole.Definition_VALUE;
  }
}
