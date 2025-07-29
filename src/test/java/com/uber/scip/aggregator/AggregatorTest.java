package com.uber.scip.aggregator;

import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mockStatic;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.spy;
import static org.mockito.Mockito.verify;

import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.util.Collections;
import java.util.Properties;
import org.apache.commons.cli.ParseException;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;

public class AggregatorTest {

  private AutoCloseable mocks;

  @TempDir public File tempFolder;
  @Mock private CommandLineConfig commandLineConfig;

  private File testJavaFile;
  private File anotherTestJavaFile;

  @BeforeEach
  public void setUp() throws IOException {
    mocks = MockitoAnnotations.openMocks(this);

    // Create a test Java file
    testJavaFile = newFile(tempFolder, "TestClass.java");
    anotherTestJavaFile = newFile(tempFolder, "AnotherTestClass.java");
    try (FileWriter writer = new FileWriter(testJavaFile)) {
      writer.write("public class TestClass { public static void main(String[] args) { } }");
    }
  }

  @Test
  public void testMainMethod_noParamsPrintsHelp() {
    assertThrows(
        IllegalArgumentException.class,
        () -> {
          try (var commandLineMockedStatic = mockStatic(CommandLineConfig.class)) {
            commandLineMockedStatic
                .when(() -> CommandLineConfig.parseArgs(any()))
                .thenThrow(new ParseException("Invalid command line arguments"));
            // Execute the main method with empty args
            Aggregator.main(new String[] {"blah-blah"});

            // Verify that printHelp was called
            commandLineMockedStatic.verify(() -> CommandLineConfig.printHelp());
          }
        });
  }

  @Test
  public void testMainMethod_withOverrides() throws IOException {
    // Create a config file
    File configFile = newFile(tempFolder, "main-test-config.properties");
    Properties props = new Properties();
    props.setProperty("classpath", tempFolder.getAbsolutePath());
    props.setProperty("semanticdb.sourceRoot", tempFolder.getAbsolutePath());
    props.setProperty("files", testJavaFile.getAbsolutePath());
    props.setProperty("output", "base_ouptut.scip");

    try (FileWriter writer = new FileWriter(configFile)) {
      props.store(writer, "Main Test Config");
    }

    // Create command line arguments
    String[] args =
        new String[] {
          "-m", configFile.getAbsolutePath(),
          "-f", anotherTestJavaFile.getAbsolutePath(),
          "-o", "output.scip"
        };

    // Create a spy of Aggregator to verify method calls
    FileAnalyzer analyzer = new FileAnalyzer();
    analyzer.withClasspath(tempFolder.getAbsolutePath());
    analyzer.withSemanticDbSourceRoot(tempFolder.getAbsolutePath());
    analyzer.withFiles(Collections.singletonList(testJavaFile.getAbsolutePath()));

    FileAnalyzer fileAnalyzerSpy = spy(analyzer);

    try (var aggregatorMockedStatic = mockStatic(Aggregator.class);
        var commandLineMockedStatic = mockStatic(CommandLineConfig.class)) {

      // Mock CommandLineConfig.parseArgs to return a config
      CommandLineConfig expectedConfig =
          new CommandLineConfig(
              configFile.getAbsolutePath(),
              "output.scip",
              Collections.singletonList(testJavaFile.getAbsolutePath()),
              null);
      commandLineMockedStatic
          .when(() -> CommandLineConfig.parseArgs(args))
          .thenReturn(expectedConfig);

      // Mock fromConfigFile to return our spy
      aggregatorMockedStatic
          .when(() -> Aggregator.fromConfigFile(configFile.getAbsolutePath(), null))
          .thenReturn(fileAnalyzerSpy);

      // Let the real main method run
      aggregatorMockedStatic.when(() -> Aggregator.main(any())).thenCallRealMethod();

      // Mock the file finding logic
      try (var mockedConfigLoader = mockStatic(ConfigLoader.class)) {
        mockedConfigLoader
            .when(() -> ConfigLoader.findJavaFilesFromPaths(any()))
            .thenReturn(Collections.singletonList(anotherTestJavaFile));

        // Execute the main method with command line args
        Aggregator.main(args);

        // Verify that analyzeFiles was called with the expected file
        verify(fileAnalyzerSpy).analyzeFiles(Collections.singletonList(anotherTestJavaFile));
        verify(fileAnalyzerSpy).withOutputPath("output.scip");
      }
    }
  }

  @Test
  public void testMainMethod_baseConfig() throws IOException {
    // Create a config file
    File configFile = newFile(tempFolder, "main-test-config.properties");
    Properties props = new Properties();
    props.setProperty("classpath", tempFolder.getAbsolutePath());
    props.setProperty("semanticdb.sourceRoot", tempFolder.getAbsolutePath());
    props.setProperty("files", testJavaFile.getAbsolutePath());
    props.setProperty("output", "base_ouptut.scip");

    try (FileWriter writer = new FileWriter(configFile)) {
      props.store(writer, "Main Test Config");
    }

    // Create command line arguments
    String[] args =
        new String[] {
          "-m", configFile.getAbsolutePath(),
        };

    // Create a spy of Aggregator to verify method calls
    FileAnalyzer analyzer = new FileAnalyzer();
    analyzer.withClasspath(tempFolder.getAbsolutePath());
    analyzer.withSemanticDbSourceRoot(tempFolder.getAbsolutePath());
    analyzer.withFiles(Collections.singletonList(testJavaFile.getAbsolutePath()));

    FileAnalyzer fileAnalyzerSpy = spy(analyzer);

    try (var aggregatorMockedStatic = mockStatic(Aggregator.class);
        var commandLineMockedStatic = mockStatic(CommandLineConfig.class)) {

      // Mock CommandLineConfig.parseArgs to return a config
      CommandLineConfig expectedConfig =
          new CommandLineConfig(configFile.getAbsolutePath(), null, null, null);
      commandLineMockedStatic
          .when(() -> CommandLineConfig.parseArgs(args))
          .thenReturn(expectedConfig);

      // Mock fromConfigFile to return our spy
      aggregatorMockedStatic
          .when(() -> Aggregator.fromConfigFile(configFile.getAbsolutePath(), null))
          .thenReturn(fileAnalyzerSpy);

      // Let the real main method run
      aggregatorMockedStatic.when(() -> Aggregator.main(any())).thenCallRealMethod();

      // Mock the file finding logic
      try (var mockedConfigLoader = mockStatic(ConfigLoader.class)) {
        mockedConfigLoader
            .when(() -> ConfigLoader.findJavaFilesFromPaths(any()))
            .thenReturn(Collections.singletonList(testJavaFile));

        // Execute the main method with command line args
        Aggregator.main(args);

        // Verify that analyzeFiles was called with the expected file
        verify(fileAnalyzerSpy).analyzeFiles(Collections.singletonList(testJavaFile));
        // Never called withOutputPath since option loader created an instance
        verify(fileAnalyzerSpy, never()).withOutputPath(any());
      }
    }
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
