package com.uber.scip.extractor;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.File;
import java.io.IOException;
import java.nio.file.Path;
import java.util.Collections;
import java.util.List;
import java.util.jar.JarEntry;
import java.util.jar.JarFile;

import com.uber.utils.TestUtils;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.mockito.MockitoAnnotations;

public class LombokJarProcessorTest {

  private AutoCloseable mocks;

  @TempDir public File tempFolder;

  private final Path fileWithLombok = TestUtils
          .getResourcePath("src/test/java/com/uber/scip/extractor/resources/jar_with_lombok.jar");
  private final Path fileWithoutLombok = TestUtils
          .getResourcePath("src/test/java/com/uber/scip/extractor/resources/jar_without_lombok.jar");

  @BeforeEach
  public void setUp() throws IOException {
    mocks = MockitoAnnotations.openMocks(this);
  }

  @Test
  public void processor_extractsLombokClasses() throws IOException {
    File outputFile = newFile(tempFolder, "output_withLombok.jar");
    LombokJarProcessor.main(new String[] {fileWithLombok.toString(), outputFile.getPath()});

    assertTrue(outputFile.exists());
    try (JarFile jarIn = new JarFile(outputFile)) {
      List<String> list =
          Collections.list(jarIn.entries()).stream().map(JarEntry::getName).toList();
      assertEquals(3, list.size());
      assertTrue(list.contains("com/uber/scip/java/LombokConfig.class"));
      assertTrue(list.contains("com/uber/scip/java/LombokConfigNested$SomeNestedClass.class"));
      assertTrue(list.contains("com/uber/scip/java/LombokConfigNested.class"));
    }
  }

  @Test
  public void processor_createsEmptyJarIfNoLombokDetected() throws IOException {
    File outputFile = newFile(tempFolder, "output_empty.jar");
    LombokJarProcessor.main(new String[] {fileWithoutLombok.toString(), outputFile.getPath()});

    assertTrue(outputFile.exists());
    try (JarFile jarIn = new JarFile(outputFile)) {
      List<JarEntry> list = Collections.list(jarIn.entries());
      assertEquals(0, list.size());
    }
  }

  @Test // Sys exit 1
  public void processor_handlesRandomFile() {
    assertThrows(
        SecurityException.class,
        () -> {
          File inputFile = newFile(tempFolder, "random.jar");
          File outputFile = newFile(tempFolder, "output_random.jar");
          LombokJarProcessor.main(new String[] {inputFile.toString(), outputFile.getPath()});
        });
  }

  @Test // Sys exit 1
  public void processor_handlesMissingFile() {
    assertThrows(
        SecurityException.class,
        () -> {
          File outputFile = newFile(tempFolder, "output_random.jar");
          LombokJarProcessor.main(new String[] {"/some/invalid/path", outputFile.getPath()});
        });
  }

  @Test // Sys exit 1
  public void processor_handlesDirInput() {
    assertThrows(
        SecurityException.class,
        () -> {
          File inputFile = newFolder(tempFolder, "dir");
          File outputFile = newFile(tempFolder, "output_random.jar");
          LombokJarProcessor.main(new String[] {inputFile.toString(), outputFile.getPath()});
        });
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

  @AfterEach
  void tearDown() throws Exception {
    mocks.close();
  }
}
