package com.uber.scip.aggregator.scip;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.mockStatic;

import com.sourcegraph.Scip;
import com.sourcegraph.scip_semanticdb.MavenPackage;
import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import javax.tools.DiagnosticCollector;
import javax.tools.JavaFileObject;

import com.uber.utils.TestUtils;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.mockito.Mock;
import org.mockito.MockedStatic;
import org.mockito.MockitoAnnotations;

public class ScipBuilderTest {

  private AutoCloseable mocks;

  @TempDir public File tempFolder;

  private ScipBuilder scipBuilder;
  private Path outputPath;
  private Path sourceRootPath;
  private List<Path> targetRoots;
  @Mock DiagnosticCollector<JavaFileObject> diagnosticCollector;

  @BeforeEach
  public void setUp() throws IOException {
    mocks = MockitoAnnotations.openMocks(this);
    scipBuilder = new ScipBuilder(diagnosticCollector);
    outputPath = newFile(tempFolder, "index.scip").toPath();

    Path resourcesPath = TestUtils.getResourcePath("src/test/java/com/uber/scip/aggregator/scip/resources");

    sourceRootPath = resourcesPath.toAbsolutePath();

    // Create target roots with sample semanticdb files
    Path targetRoot = newFolder(tempFolder, "targetroot").toPath();
    targetRoots = List.of(sourceRootPath);
  }

  @Test
  public void testBuildScip() throws IOException {
    BuildOptions options =
        new BuildOptions.Builder()
            .withTargetRoots(targetRoots)
            .withOutputPath(outputPath)
            .withSourceRoot(sourceRootPath)
            .withClasspathString("")
            .build();

    scipBuilder.buildScip(options);

    // verify index.scip was created in the output path
    assertTrue(Files.exists(outputPath));

    Scip.Index index = Scip.Index.parseFrom(Files.newInputStream(outputPath));
    assertEquals("java-code", index.getMetadata().getProjectRoot());
    assertEquals(1, index.getDocumentsCount());
    assertTrue(
        index
            .getDocuments(0)
            .getRelativePath()
            .endsWith("java-sample/src/main/java/com/uber/scip/java/App.java"));
  }

  @Test
  public void testCollectMavenPackagesFromClasspath_EmptyClasspath() {
    String classpath = "";
    List<MavenPackage> packages = scipBuilder.collectMavenPackagesFromClasspath(classpath);
    assertTrue(packages.isEmpty(), "Empty classpath should return empty list");
  }

  @Test
  public void testCollectMavenPackagesFromClasspath_NullClasspath() {
    List<MavenPackage> packages = scipBuilder.collectMavenPackagesFromClasspath(null);
    assertTrue(packages.isEmpty(), "Null classpath should return empty list");
  }

  @Test
  public void testCollectMavenPackagesFromClasspath_WithJars() throws IOException {
    // Create test JAR files in temp directory
    MockedStatic<UberScipWriter> mockedStatic = mockStatic(UberScipWriter.class);

    Path jar1 = newFile(tempFolder, "com.example.artifact-1.0.0.jar").toPath();
    Path jar2 = newFile(tempFolder, "simple-lib-2.1.jar").toPath();
    Path nonJar = newFile(tempFolder, "not-a-jar.txt").toPath();
    mockedStatic
        .when(() -> UberScipWriter.is3rdPartyOutputPath(eq(jar1.toString())))
        .thenReturn(true);

    String classpath =
        String.join(
            System.getProperty("path.separator"),
            jar1.toString(),
            jar2.toString(),
            nonJar.toString());

    List<MavenPackage> packages = scipBuilder.collectMavenPackagesFromClasspath(classpath);

    assertEquals(1, packages.size(), "Should find 1 Maven package");

    // Verify first package
    MavenPackage pkg1 = packages.get(0);
    assertEquals(".", pkg1.groupId);
    assertEquals(".", pkg1.artifactId);
    assertEquals("com.example.artifact-1.0.0", pkg1.version());
  }

  @Test
  public void testExtractMavenInfoFromJar_SimpleFormat() throws IOException {
    Path jarPath = newFile(tempFolder, "artifact-1.0.0.jar").toPath();
    MavenPackage pkg = scipBuilder.extractMavenInfoFromJar(jarPath);

    assertEquals(".", pkg.groupId);
    assertEquals(".", pkg.artifactId);
    assertEquals("artifact-1.0.0", pkg.version());
  }

  @Test
  public void testExtractMavenInfoFromJar_RandomFormat() throws IOException {
    Path jarPath = newFile(tempFolder, "invalid_format.jar").toPath();
    MavenPackage pkg = scipBuilder.extractMavenInfoFromJar(jarPath);

    assertEquals(".", pkg.groupId);
    assertEquals(".", pkg.artifactId);
    assertEquals("invalid_format", pkg.version());
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
