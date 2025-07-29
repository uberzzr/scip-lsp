package com.uber.scip.aggregator.scip;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.anyBoolean;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.spy;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import com.sourcegraph.Scip;
import com.sourcegraph.scip_semanticdb.ScipSemanticdbOptions;
import java.io.IOException;
import java.util.List;
import javax.tools.Diagnostic;
import javax.tools.DiagnosticCollector;
import javax.tools.JavaFileObject;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;

public class UberScipWriterTest {

  @Mock DiagnosticCollector<JavaFileObject> diagnosticCollector;

  @Mock Diagnostic diagnostic;
  @Mock JavaFileObject fileObject;

  @BeforeEach
  public void setUp() throws IOException {
    MockitoAnnotations.initMocks(this);
  }

  @Test
  public void testEmitTypedSortsDocuments() throws IOException {
    // Create mock options
    ScipSemanticdbOptions options = mock(ScipSemanticdbOptions.class);

    // Create test documents with different paths
    Scip.Document doc1 = Scip.Document.newBuilder().setRelativePath("z/file.java").build();
    Scip.Document doc2 = Scip.Document.newBuilder().setRelativePath("a/file.java").build();

    // Create test metadata
    Scip.Metadata metadata = Scip.Metadata.newBuilder().build();

    // Create test index with unsorted documents
    Scip.Index originalIndex =
        Scip.Index.newBuilder().addDocuments(doc1).addDocuments(doc2).setMetadata(metadata).build();

    BuildOptions buildOptions = BuildOptions.defaultOptions();
    UberScipWriter writer = spy(new UberScipWriter(options, buildOptions, diagnosticCollector));

    // Call the method
    writer.emitTyped(originalIndex);

    // Verify originalEmitTyped was called with the sorted index
    org.mockito.ArgumentCaptor<Scip.Index> indexCaptor =
        org.mockito.ArgumentCaptor.forClass(Scip.Index.class);
    verify(writer).originalEmitTyped(indexCaptor.capture());

    // Verify the documents are sorted by relative path
    Scip.Index processedIndex = indexCaptor.getValue();
    List<Scip.Document> sortedDocs = processedIndex.getDocumentsList();
    assertEquals("a/file.java", sortedDocs.get(0).getRelativePath());
    assertEquals("z/file.java", sortedDocs.get(1).getRelativePath());

    // Verify project root is set correctly
    assertEquals("java-code", processedIndex.getMetadata().getProjectRoot());
  }

  @Test
  public void testEmitTypedWritesDiagnostics() throws IOException {
    // Create mock options
    ScipSemanticdbOptions options = mock(ScipSemanticdbOptions.class);

    // Create test documents with different paths
    Scip.Document doc1 = Scip.Document.newBuilder().setRelativePath("z/file.java").build();
    Scip.Document doc2 =
        Scip.Document.newBuilder()
            .setRelativePath("bazel-out/host/bin/3rdparty/some/file.java")
            .build();

    // Create test metadata
    Scip.Metadata metadata = Scip.Metadata.newBuilder().build();
    when(diagnosticCollector.getDiagnostics()).thenReturn(List.of(diagnostic));
    when(diagnostic.getSource()).thenReturn(fileObject);
    when(fileObject.getName()).thenReturn("z/file.java");
    when(fileObject.getCharContent(anyBoolean())).thenReturn("Java content");
    when(diagnostic.getKind()).thenReturn(Diagnostic.Kind.ERROR);
    when(diagnostic.getMessage(null)).thenReturn("Error message");
    when(diagnostic.getLineNumber()).thenReturn(1L);
    when(diagnostic.getColumnNumber()).thenReturn(1L);
    when(diagnostic.getStartPosition()).thenReturn(1L);
    when(diagnostic.getEndPosition()).thenReturn(1L);

    // Create test index with unsorted documents
    Scip.Index originalIndex =
        Scip.Index.newBuilder().addDocuments(doc1).addDocuments(doc2).setMetadata(metadata).build();

    BuildOptions buildOptions = BuildOptions.defaultOptions();
    UberScipWriter writer = spy(new UberScipWriter(options, buildOptions, diagnosticCollector));

    // Call the method
    writer.emitTyped(originalIndex);

    // Verify originalEmitTyped was called with the sorted index
    org.mockito.ArgumentCaptor<Scip.Index> indexCaptor =
        org.mockito.ArgumentCaptor.forClass(Scip.Index.class);
    verify(writer).originalEmitTyped(indexCaptor.capture());

    // Verify the documents are sorted by relative path
    Scip.Index processedIndex = indexCaptor.getValue();
    List<Scip.Document> sortedDocs = processedIndex.getDocumentsList();
    assertEquals("bazel-out/host/bin/3rdparty/some/file.java", sortedDocs.get(0).getRelativePath());
    assertEquals("z/file.java", sortedDocs.get(1).getRelativePath());
    assertTrue(sortedDocs.get(0).getOccurrencesList().isEmpty());
    assertEquals(1, sortedDocs.get(1).getOccurrencesCount());
    assertEquals(1, sortedDocs.get(1).getOccurrencesList().get(0).getDiagnosticsCount());
    assertEquals(
        "Error message",
        sortedDocs.get(1).getOccurrencesList().get(0).getDiagnostics(0).getMessage());
  }

  @Test
  public void testShading3rdPartySymbols() throws IOException {
    ScipSemanticdbOptions options = mock(ScipSemanticdbOptions.class);

    // Create test documents with 3rd party path and occurrences
    Scip.Occurrence def =
        Scip.Occurrence.newBuilder()
            .setSymbol("some.symbol")
            .setSymbolRoles(Scip.SymbolRole.Definition_VALUE)
            .build();
    Scip.Occurrence usage =
        Scip.Occurrence.newBuilder()
            .setSymbol("some.symbol")
            .setSymbolRoles(Scip.SymbolRole.Import_VALUE)
            .build();

    Scip.Document thirdPartyDoc =
        Scip.Document.newBuilder()
            .setRelativePath("bazel-out/host/bin/3rdparty/some/file.java")
            .addOccurrences(def)
            .addOccurrences(usage)
            .build();

    Scip.Index originalIndex =
        Scip.Index.newBuilder()
            .addDocuments(thirdPartyDoc)
            .setMetadata(Scip.Metadata.newBuilder().build())
            .build();

    // Test with shading enabled
    BuildOptions shadingEnabled =
        new BuildOptions.Builder().shouldShade3rdPartySymbols(true).build();
    UberScipWriter writerWithShading =
        spy(new UberScipWriter(options, shadingEnabled, diagnosticCollector));
    writerWithShading.emitTyped(originalIndex);

    // Verify document was processed with shading
    org.mockito.ArgumentCaptor<Scip.Index> shadingCaptor =
        org.mockito.ArgumentCaptor.forClass(Scip.Index.class);
    verify(writerWithShading).originalEmitTyped(shadingCaptor.capture());
    assertEquals(1, shadingCaptor.getValue().getDocuments(0).getOccurrencesCount());

    // Test with shading disabled
    BuildOptions shadingDisabled =
        new BuildOptions.Builder().shouldShade3rdPartySymbols(false).build();
    UberScipWriter writerNoShading =
        spy(new UberScipWriter(options, shadingDisabled, diagnosticCollector));
    writerNoShading.emitTyped(originalIndex);

    // Verify document was preserved without shading
    org.mockito.ArgumentCaptor<Scip.Index> noShadingCaptor =
        org.mockito.ArgumentCaptor.forClass(Scip.Index.class);
    verify(writerNoShading).originalEmitTyped(noShadingCaptor.capture());
    assertEquals(2, noShadingCaptor.getValue().getDocuments(0).getOccurrencesCount());
  }

  @Test
  public void testShadingIdlSymbols() throws IOException {
    ScipSemanticdbOptions options = mock(ScipSemanticdbOptions.class);

    // Create test documents with IDL path and occurrences
    Scip.Occurrence def =
        Scip.Occurrence.newBuilder()
            .setSymbol("some.idl.symbol")
            .setSymbolRoles(Scip.SymbolRole.Definition_VALUE)
            .build();
    Scip.Occurrence usage =
        Scip.Occurrence.newBuilder()
            .setSymbol("some.symbol")
            .setSymbolRoles(Scip.SymbolRole.Import_VALUE)
            .build();

    Scip.Document idlDoc =
        Scip.Document.newBuilder()
            .setRelativePath("bazel-out/host/bin/idl/some/file.proto")
            .addOccurrences(def)
            .addOccurrences(usage)
            .build();

    Scip.Index originalIndex =
        Scip.Index.newBuilder()
            .addDocuments(idlDoc)
            .setMetadata(Scip.Metadata.newBuilder().build())
            .build();

    // Test with shading enabled
    BuildOptions shadingEnabled = new BuildOptions.Builder().shouldShadeIdlSymbols(true).build();
    UberScipWriter writerWithShading =
        spy(new UberScipWriter(options, shadingEnabled, diagnosticCollector));
    writerWithShading.emitTyped(originalIndex);

    // Verify document was processed with shading
    org.mockito.ArgumentCaptor<Scip.Index> shadingCaptor =
        org.mockito.ArgumentCaptor.forClass(Scip.Index.class);
    verify(writerWithShading).originalEmitTyped(shadingCaptor.capture());
    assertEquals(1, shadingCaptor.getValue().getDocuments(0).getOccurrencesCount());

    // Test with shading disabled
    BuildOptions shadingDisabled = new BuildOptions.Builder().shouldShadeIdlSymbols(false).build();
    UberScipWriter writerNoShading =
        spy(new UberScipWriter(options, shadingDisabled, diagnosticCollector));
    writerNoShading.emitTyped(originalIndex);

    // Verify document was preserved without shading
    org.mockito.ArgumentCaptor<Scip.Index> noShadingCaptor =
        org.mockito.ArgumentCaptor.forClass(Scip.Index.class);
    verify(writerNoShading).originalEmitTyped(noShadingCaptor.capture());
    assertEquals(2, noShadingCaptor.getValue().getDocuments(0).getOccurrencesCount());
  }
}
