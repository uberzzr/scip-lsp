package com.uber.scip.aggregator;

import com.sun.source.util.JavacTask;
import com.uber.scip.aggregator.scip.BuildOptions;
import com.uber.scip.aggregator.scip.CompilationIssue;
import com.uber.scip.aggregator.scip.ScipBuilder;
import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import java.util.function.Supplier;
import java.util.stream.Collectors;
import java.util.stream.StreamSupport;
import javax.tools.DiagnosticCollector;
import javax.tools.JavaCompiler;
import javax.tools.JavaFileObject;
import javax.tools.StandardJavaFileManager;
import javax.tools.ToolProvider;
import org.apache.commons.io.FileUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/** A tool that uses the JavaC Compiler API to analyze Java files. */
public class FileAnalyzer {
  private static final Logger logger = LoggerFactory.getLogger(FileAnalyzer.class);

  private final JavaCompiler compiler;
  private final StandardJavaFileManager fileManager;
  private final CompilerOptions compilerOptions;
  private final SemanticDbManager semanticDbManager;
  private final List<String> filePaths;
  private final Supplier<DiagnosticCollector<JavaFileObject>> diagnosticCollectorSupplier;
  private String outputPath = "index.scip";

  public FileAnalyzer() {
    this(
        ToolProvider.getSystemJavaCompiler(),
        ToolProvider.getSystemJavaCompiler().getStandardFileManager(null, null, null),
        new CompilerOptions(),
        new SemanticDbManager(),
        DiagnosticCollector::new);
  }

  public FileAnalyzer(
      JavaCompiler compiler,
      StandardJavaFileManager fileManager,
      CompilerOptions compilerOptions,
      SemanticDbManager semanticDbManager,
      Supplier<DiagnosticCollector<JavaFileObject>> collectorSupplier) {
    this.compiler = compiler;
    this.fileManager = fileManager;
    this.compilerOptions = compilerOptions;
    this.semanticDbManager = semanticDbManager;
    this.filePaths = new ArrayList<>();
    this.diagnosticCollectorSupplier = collectorSupplier;
  }

  /** Adds compiler options */
  public void withOptions(List<String> options) {
    this.compilerOptions.addOptions(options);
  }

  /** Sets the classpath for compilation */
  public void withClasspath(String classpath) {
    this.compilerOptions.setClasspath(classpath);
  }

  /** Sets the semanticdb source root */
  public void withSemanticDbSourceRoot(String sourceRoot) {
    this.semanticDbManager.setSourceRoot(sourceRoot);
  }

  /** Sets the semanticdb target output root */
  public void withSemanticDbTargetRoot(String targetRoot) {
    this.semanticDbManager.setTargetRoot(targetRoot);
  }

  /** Sets the semanticdb target output root */
  public void withSemanticDbPlugin(String plugin) {
    this.semanticDbManager.setPlugin(plugin);
  }

  /** Adds files to be analyzed */
  public void withFiles(List<String> files) {
    this.filePaths.addAll(files);
  }

  /** Specify output file. */
  public void withOutputPath(String output) {
    this.outputPath = output;
  }

  /** Gets the list of file paths to be analyzed */
  public List<String> getFilePaths() {
    return filePaths;
  }

  public JavaCompiler getCompiler() {
    return compiler;
  }

  public CompilerOptions getCompilerOptions() {
    return compilerOptions;
  }

  public StandardJavaFileManager getFileManager() {
    return fileManager;
  }

  public SemanticDbManager getSemanticDbManager() {
    return semanticDbManager;
  }

  public String getOutputPath() {
    return outputPath;
  }

  /** Analyzes the provided Java files using the JavaC API */
  public void analyzeFiles(List<File> files) throws IOException {
    // Parse and filter files first
    Iterable<? extends JavaFileObject> compilationUnits =
        fileManager.getJavaFileObjectsFromFiles(files);
    Path outputDir = Files.createTempDirectory(UUID.randomUUID().toString());
    Path scipOutputPath = Paths.get(this.outputPath);
    DiagnosticCollector<JavaFileObject> diagnosticCollector =
        this.diagnosticCollectorSupplier.get();

    if ((int) StreamSupport.stream(compilationUnits.spliterator(), false).count() < 1) {
      logger.debug("No valid files to analyze. Writing empty index.");
      // Write empty index file to cache output.
      buildScip(outputDir, scipOutputPath, diagnosticCollector);
      return;
    }
    List<String> allOptions = compilerOptions.getCompilerOptions();

    try {
      semanticDbManager.setTargetRoot(outputDir.toString());
      allOptions.add(semanticDbManager.formatSemanticDBPluginConfig());
      runAnalysis(diagnosticCollector, allOptions, compilationUnits);
    } catch (IllegalStateException e) {
      // Sometimes javac will fail due to incomplete compile cycle.
      // In this case, we can try to limit symbol generation to given target.
      // Example: //3rdparty/jvm/org/apache/hadoop:hadoop-yarn-common-2.8.2.jar
      // Sometimes target missing required dependencies.
      // Example //3rdparty/jvm/org/springframework/boot:spring-boot-2.7.18.jar missing
      // 3rdparty/jvm/jakarta/validation:jakarta.validation-api-2.0.2.jar
      // In this case internal symbols will be unresolved.
      diagnosticCollector = this.diagnosticCollectorSupplier.get();
      compilerOptions.setClasspath(semanticDbManager.getPlugin());
      allOptions = compilerOptions.getCompilerOptions();
      allOptions.add(semanticDbManager.formatSemanticDBPluginConfig());
      try {
        runAnalysis(diagnosticCollector, allOptions, compilationUnits);
      } catch (IllegalStateException nestedStateException) {
        // Best effort
        // Could result java.lang.IllegalStateException: java.lang.NullPointerException: Cannot read
        // field "tree" because "env" is null
        logger.debug("Nested state exception: {}", e.getMessage());
      }
    } finally {
      try {
        buildScip(outputDir, scipOutputPath, diagnosticCollector);
        FileUtils.deleteDirectory(outputDir.toFile());
      } catch (IOException e) {
        logger.warn("Failed to delete temporary directory: {}", outputDir, e);
      }
    }

    // Report any issues that occurred during analysis
    List<CompilationIssue> analysisIssues =
        diagnosticCollector.getDiagnostics().stream()
            .map(CompilationIssue::new)
            .collect(Collectors.toList());

    if (!analysisIssues.isEmpty()) {
      logger.debug("Analysis produced {} issues:", analysisIssues.size());
      analysisIssues.forEach(
          issue ->
              logger.debug(
                  "  {} in {} at position {}:{} - {}",
                  issue.getKind(),
                  issue.getSource(),
                  issue.getLineNumber(),
                  issue.getColumnNumberStart(),
                  issue.getMessage()));
    }
  }

  private void runAnalysis(
      DiagnosticCollector<JavaFileObject> diagnosticCollector,
      List<String> allOptions,
      Iterable<? extends JavaFileObject> compilationUnits)
      throws IOException {
    JavacTask task = getJavacTask(diagnosticCollector, allOptions, compilationUnits);

    logger.debug("Starting analysis...");
    task.analyze();
    logger.debug("Analysis complete. Building SCIP index...");
  }

  private JavacTask getJavacTask(
      DiagnosticCollector<JavaFileObject> diagnosticCollector,
      List<String> allOptions,
      Iterable<? extends JavaFileObject> compilationUnits) {
    return (JavacTask)
        compiler.getTask(
            null, // Writer for compiler output
            fileManager, // File manager
            diagnosticCollector, // Diagnostic listener
            allOptions, // Compiler options
            null, // Classes to be processed by annotation processors
            compilationUnits // Compilation units
            );
  }

  private void buildScip(
      Path outputDir, Path scipOutputPath, DiagnosticCollector<JavaFileObject> diagnosticCollector)
      throws IOException {
    BuildOptions options =
        new BuildOptions.Builder()
            .withClasspathString(this.compilerOptions.getClasspath())
            .withTargetRoots(List.of(outputDir))
            .withSourceRoot(Paths.get(semanticDbManager.getSourceRoot()))
            .withOutputPath(scipOutputPath)
            .build();
    ScipBuilder build = new ScipBuilder(diagnosticCollector);
    build.buildScip(options);
    logger.debug("SCIP index generation complete: {}", scipOutputPath);
  }
}
