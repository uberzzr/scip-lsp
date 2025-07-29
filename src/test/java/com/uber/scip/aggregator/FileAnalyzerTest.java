package com.uber.scip.aggregator;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.ArgumentMatchers.isNull;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import com.sun.source.util.JavacTask;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Properties;
import javax.tools.Diagnostic;
import javax.tools.DiagnosticCollector;
import javax.tools.JavaCompiler;
import javax.tools.JavaFileObject;
import javax.tools.StandardJavaFileManager;
import org.assertj.core.api.Assertions;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;

public class FileAnalyzerTest {

  private AutoCloseable mocks;

  @TempDir public File tempFolder;

  @Mock private JavaCompiler mockCompiler;

  @Mock private StandardJavaFileManager mockFileManager;

  @Mock private JavacTask mockJavacTask;
  @Mock private DiagnosticCollector mockDiagnosticCollector;
  @Mock private Diagnostic diagnostic;

  private FileAnalyzer fileAnalyzer;
  private File testJavaFile;

  @BeforeEach
  public void setUp() throws IOException {
    mocks = MockitoAnnotations.openMocks(this);

    fileAnalyzer =
        new FileAnalyzer(
            this.mockCompiler,
            this.mockFileManager,
            new CompilerOptions(),
            new SemanticDbManager(),
            () -> mockDiagnosticCollector);

    // Create a test Java file
    testJavaFile = newFile(tempFolder, "TestClass.java");
    try (FileWriter writer = new FileWriter(testJavaFile)) {
      writer.write("public class TestClass { public static void main(String[] args) { } }");
    }
  }

  @Test
  public void testAnalyzeFiles() throws IOException {
    // Setup compiler task mock
    when(mockCompiler.getTask(isNull(), eq(mockFileManager), any(), any(), isNull(), any()))
        .thenReturn(mockJavacTask);

    // Setup file manager mock to return our test file
    List<File> files = Collections.singletonList(testJavaFile);
    JavaFileObject mockJavaFileObject = mock(JavaFileObject.class);
    ArrayList list = new ArrayList<>();
    list.add(mockJavaFileObject);
    when(mockFileManager.getJavaFileObjectsFromFiles(any())).thenReturn(list);

    // Configure semanticdb manager
    this.fileAnalyzer.withSemanticDbSourceRoot(tempFolder.getAbsolutePath());

    // Act
    this.fileAnalyzer.analyzeFiles(files);

    // Verify
    verify(mockJavacTask).analyze();
  }

  @Test
  public void testAnalyzeFiles_withErrors() throws IOException {
    // Setup compiler task mock
    when(mockCompiler.getTask(isNull(), eq(mockFileManager), any(), any(), isNull(), any()))
        .thenReturn(mockJavacTask);
    when(mockDiagnosticCollector.getDiagnostics()).thenReturn(List.of(diagnostic));

    // Setup file manager mock to return our test file
    List<File> files = Collections.singletonList(testJavaFile);
    JavaFileObject mockJavaFileObject = mock(JavaFileObject.class);
    ArrayList list = new ArrayList<>();
    list.add(mockJavaFileObject);
    when(mockFileManager.getJavaFileObjectsFromFiles(any())).thenReturn(list);

    // Configure semanticdb manager
    this.fileAnalyzer.withSemanticDbSourceRoot(tempFolder.getAbsolutePath());

    // Act
    this.fileAnalyzer.analyzeFiles(files);

    // Verify
    verify(mockJavacTask).analyze();
  }

  @Test
  public void testAnalyzeFiles_fallbackToEmptyClasspath() throws IOException {
    ArrayList<String> classpath = new ArrayList<>();
    classpath.add("some_classpath");
    classpath.add("some_other_classpath");
    classpath.add("path_to_plugin");
    when(mockCompiler.getTask(isNull(), eq(mockFileManager), any(), any(), isNull(), any()))
        .thenReturn(mockJavacTask);
    when(mockJavacTask.analyze())
        .thenThrow(new IllegalStateException("Test exception"))
        .thenReturn(null);

    // Setup file manager mock to return our test file
    List<File> files = Collections.singletonList(testJavaFile);
    JavaFileObject mockJavaFileObject = mock(JavaFileObject.class);
    ArrayList list = new ArrayList<>();
    list.add(mockJavaFileObject);
    when(mockFileManager.getJavaFileObjectsFromFiles(any())).thenReturn(list);

    // Configure semanticdb manager
    this.fileAnalyzer.withSemanticDbSourceRoot(tempFolder.getAbsolutePath());
    this.fileAnalyzer.withSemanticDbPlugin("path_to_plugin");
    this.fileAnalyzer.withClasspath(String.join(":", classpath));

    // Act
    this.fileAnalyzer.analyzeFiles(files);

    // Verify
    verify(mockJavacTask, times(2)).analyze();
    ArgumentCaptor<List<String>> optionsCaptor = ArgumentCaptor.forClass(List.class);

    verify(mockCompiler, times(2))
        .getTask(
            isNull(), // Writer
            eq(mockFileManager), // FileManager
            any(), // DiagnosticListener
            optionsCaptor.capture(), // Options
            isNull(), // Classes
            any() // CompilationUnits
            );

    // Verify the compiler options - should be empty on fallback
    List<List<String>> capturedOptions = optionsCaptor.getAllValues();
    List<String> firstCall = capturedOptions.get(0);
    List<String> secondCall = capturedOptions.get(1);

    List<String> firstExpectedOptions = Arrays.asList("-classpath", String.join(":", classpath));
    List<String> secondExpectedOptions = Arrays.asList("-classpath", "path_to_plugin");

    assertTrue(Collections.indexOfSubList(firstCall, firstExpectedOptions) > 0);
    assertTrue(Collections.indexOfSubList(firstCall, secondExpectedOptions) < 0);
    assertTrue(Collections.indexOfSubList(secondCall, firstExpectedOptions) < 0);
    assertTrue(Collections.indexOfSubList(secondCall, secondExpectedOptions) > 0);
  }

  @Test
  public void testAnalyzeFiles_nestedStateException() throws IOException {
    ArrayList<String> classpath = new ArrayList<>();
    classpath.add("some_classpath");
    classpath.add("some_other_classpath");
    classpath.add("path_to_plugin");
    when(mockCompiler.getTask(isNull(), eq(mockFileManager), any(), any(), isNull(), any()))
        .thenReturn(mockJavacTask);
    when(mockJavacTask.analyze())
        .thenThrow(new IllegalStateException("Test exception"))
        .thenThrow(new IllegalStateException("Nested Test exception"));

    // Setup file manager mock to return our test file
    List<File> files = Collections.singletonList(testJavaFile);
    JavaFileObject mockJavaFileObject = mock(JavaFileObject.class);
    ArrayList list = new ArrayList<>();
    list.add(mockJavaFileObject);
    when(mockFileManager.getJavaFileObjectsFromFiles(any())).thenReturn(list);

    // Configure semanticdb manager
    this.fileAnalyzer.withSemanticDbSourceRoot(tempFolder.getAbsolutePath());
    this.fileAnalyzer.withSemanticDbPlugin("path_to_plugin");
    this.fileAnalyzer.withClasspath(String.join(":", classpath));

    // Act
    this.fileAnalyzer.analyzeFiles(files);

    // Verify
    verify(mockJavacTask, times(2)).analyze();
    ArgumentCaptor<List<String>> optionsCaptor = ArgumentCaptor.forClass(List.class);

    verify(mockCompiler, times(2))
        .getTask(
            isNull(), // Writer
            eq(mockFileManager), // FileManager
            any(), // DiagnosticListener
            optionsCaptor.capture(), // Options
            isNull(), // Classes
            any() // CompilationUnits
            );

    // Verify the compiler options - should be empty on fallback
    List<List<String>> capturedOptions = optionsCaptor.getAllValues();
    List<String> firstCall = capturedOptions.get(0);
    List<String> secondCall = capturedOptions.get(1);

    List<String> firstExpectedOptions = Arrays.asList("-classpath", String.join(":", classpath));
    List<String> secondExpectedOptions = Arrays.asList("-classpath", "path_to_plugin");

    assertTrue(Collections.indexOfSubList(firstCall, firstExpectedOptions) > 0);
    assertTrue(Collections.indexOfSubList(firstCall, secondExpectedOptions) < 0);
    assertTrue(Collections.indexOfSubList(secondCall, firstExpectedOptions) < 0);
    assertTrue(Collections.indexOfSubList(secondCall, secondExpectedOptions) > 0);
  }

  @Test
  public void testFromConfigFile() throws IOException {
    // Create a config file
    File configFile = newFile(tempFolder, "test-config.properties");
    Properties props = new Properties();
    props.setProperty("classpath", "/test/classpath");
    props.setProperty("options", "-Xlint:all,-deprecation");
    props.setProperty("semanticdb.sourceRoot", "/test/source/root");
    props.setProperty("semanticdb.targetRoot", "/test/target/root");
    props.setProperty("files", "File1.java,File2.java");

    try (FileWriter writer = new FileWriter(configFile)) {
      props.store(writer, "Test Config");
    }

    // Act
    FileAnalyzer result = Aggregator.fromConfigFile(configFile.getAbsolutePath(), null);

    // Assert
    assertEquals("/test/classpath", result.getCompilerOptions().getClasspath());
    Assertions.assertThat(result.getCompilerOptions().getOptions())
        .containsAnyOf("-Xlint:all", "-deprecation");
    assertEquals("/test/source/root", result.getSemanticDbManager().getSourceRoot());
    assertEquals("/test/target/root", result.getSemanticDbManager().getTargetRoot());
    assertEquals(Arrays.asList("File1.java", "File2.java"), result.getFilePaths());
  }

  private static File newFile(File parent, String child) throws IOException {
    File result = new File(parent, child);
    result.createNewFile();
    return result;
  }

  @AfterEach
  void tearDown() throws Exception {
    mocks.close();
  }
}
