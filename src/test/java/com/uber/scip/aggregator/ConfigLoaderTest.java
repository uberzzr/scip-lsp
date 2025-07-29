package com.uber.scip.aggregator;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.verifyNoMoreInteractions;

import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.util.Arrays;
import java.util.List;
import java.util.Properties;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.mockito.MockedConstruction;
import org.mockito.Mockito;

public class ConfigLoaderTest {

  @TempDir public File tempFolder;

  private File configFile;
  private File javaFile1;
  private File javaFile2;
  private File nonJavaFile;
  private File subDir;
  private File javaFileInSubDir;

  @BeforeEach
  public void setUp() throws IOException {
    // Create a config file for testing
    configFile = newFile(tempFolder, "test-config.properties");

    // Create test Java files and directories
    javaFile1 = newFile(tempFolder, "Test1.java");
    javaFile2 = newFile(tempFolder, "Test2.java");
    nonJavaFile = newFile(tempFolder, "notJava.txt");
    subDir = newFolder(tempFolder, "subdir");
    javaFileInSubDir = new File(subDir, "SubTest.java");
    javaFileInSubDir.createNewFile();
  }

  @Test
  public void testLoadFromFile_explicitFiles() throws IOException {
    // Write test properties to config file
    Properties props = new Properties();
    props.setProperty("classpath", "/path/to/classes:/another/path");
    props.setProperty("options", "-Xmx2g,-verbose");
    props.setProperty("semanticdb.sourceRoot", "/source/root");
    props.setProperty("semanticdb.targetRoot", "/target/root");
    props.setProperty("files", "file1.java,file2.java");
    props.setProperty("output", "test_output.scip");

    try (FileWriter writer = new FileWriter(configFile)) {
      props.store(writer, "Test properties");
    }

    try (MockedConstruction<FileAnalyzer> mockedConstruction =
        Mockito.mockConstruction(FileAnalyzer.class)) {
      ConfigLoader.loadFromFile(configFile.getAbsolutePath(), null);

      // Get the constructed mock instance
      FileAnalyzer analyzer = mockedConstruction.constructed().get(0);

      // Verify the correct methods were called with expected arguments
      verify(analyzer).withClasspath("/path/to/classes:/another/path");
      verify(analyzer).withOptions(Arrays.asList("-Xmx2g", "-verbose"));
      verify(analyzer).withSemanticDbSourceRoot("/source/root");
      verify(analyzer).withSemanticDbTargetRoot("/target/root");
      verify(analyzer).withFiles(Arrays.asList("file1.java", "file2.java"));
      verify(analyzer).withOutputPath("test_output.scip");
    }
  }

  @Test
  public void testLoadFromFile_fileWithSources() throws IOException {
    // Write source list to file
    File sourceListFile = newFile(tempFolder, "sourceList.txt");
    try (FileWriter writer = new FileWriter(sourceListFile)) {
      writer.write(javaFile1.getName() + "\n");
      writer.write(javaFile2.getName() + "\n");
    }

    // Write test properties to config file
    Properties props = new Properties();
    props.setProperty("classpath", "/path/to/classes:/another/path");
    props.setProperty("options", "-Xmx2g,-verbose");
    props.setProperty("semanticdb.sourceRoot", "/source/root");
    props.setProperty("semanticdb.targetRoot", "/target/root");
    props.setProperty("files_file", sourceListFile.getAbsolutePath());

    try (FileWriter writer = new FileWriter(configFile)) {
      props.store(writer, "Test properties");
    }

    try (MockedConstruction<FileAnalyzer> mockedConstruction =
        Mockito.mockConstruction(FileAnalyzer.class)) {
      ConfigLoader.loadFromFile(configFile.getAbsolutePath(), null);

      // Get the constructed mock instance
      FileAnalyzer analyzer = mockedConstruction.constructed().get(0);

      // Verify the correct methods were called with expected arguments
      verify(analyzer).withClasspath("/path/to/classes:/another/path");
      verify(analyzer).withOptions(Arrays.asList("-Xmx2g", "-verbose"));
      verify(analyzer).withSemanticDbSourceRoot("/source/root");
      verify(analyzer).withSemanticDbTargetRoot("/target/root");
      verify(analyzer).withFiles(Arrays.asList("Test1.java", "Test2.java"));
    }
  }

  @Test
  public void testLoadFromFileWithMissingProperties() throws IOException {
    // Write minimal properties to config file
    Properties props = new Properties();
    props.setProperty("classpath", "/path/to/classes");
    // Intentionally omit other properties

    try (FileWriter writer = new FileWriter(configFile)) {
      props.store(writer, "Minimal test properties");
    }

    // Load the config
    try (MockedConstruction<FileAnalyzer> mockedConstruction =
        Mockito.mockConstruction(FileAnalyzer.class)) {
      ConfigLoader.loadFromFile(configFile.getAbsolutePath(), tempFolder.getAbsolutePath());

      // Get the constructed mock instance
      FileAnalyzer analyzer = mockedConstruction.constructed().get(0);

      // Verify the correct methods were called with expected arguments
      verify(analyzer).withClasspath("/path/to/classes");
      verifyNoMoreInteractions(analyzer);
    }
  }

  @Test
  public void testFindJavaFilesFromPaths() {
    List<String> paths =
        Arrays.asList(
            javaFile1.getAbsolutePath(),
            javaFile2.getAbsolutePath(),
            nonJavaFile.getAbsolutePath(),
            subDir.getAbsolutePath());

    List<File> javaFiles = ConfigLoader.findJavaFilesFromPaths(paths);

    // Should find 3 Java files: javaFile1, javaFile2, and javaFileInSubDir
    assertEquals(3, javaFiles.size());
    assertTrue(javaFiles.contains(javaFile1));
    assertTrue(javaFiles.contains(javaFile2));
    assertTrue(javaFiles.contains(javaFileInSubDir));
    assertFalse(javaFiles.contains(nonJavaFile));
  }

  @Test
  public void testFindJavaFilesFromEmptyPaths() {
    List<String> emptyPaths = Arrays.asList();
    List<File> javaFiles = ConfigLoader.findJavaFilesFromPaths(emptyPaths);
    assertTrue(javaFiles.isEmpty());
  }

  private static File newFile(File parent, String child) throws IOException {
    File result = new File(parent, child);
    result.createNewFile();
    return result;
  }

  private static File newFolder(File root, String... subDirs) throws IOException {
    String subFolder = String.join("/", subDirs);
    File result = new File(root, subFolder);
    if (!result.mkdirs()) {
      throw new IOException("Couldn't create folders " + root);
    }
    return result;
  }
}
