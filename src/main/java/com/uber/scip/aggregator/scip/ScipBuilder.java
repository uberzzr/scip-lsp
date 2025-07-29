package com.uber.scip.aggregator.scip;

import static com.uber.scip.aggregator.scip.UberScipWriter.is3rdPartyOutputPath;

import com.sourcegraph.lsif_protocol.LsifToolInfo;
import com.sourcegraph.scip_semanticdb.MavenPackage;
import com.sourcegraph.scip_semanticdb.ScipOutputFormat;
import com.sourcegraph.scip_semanticdb.ScipSemanticdb;
import com.sourcegraph.scip_semanticdb.ScipSemanticdbOptions;
import com.sourcegraph.scip_semanticdb.ScipSemanticdbReporter;
import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import javax.tools.DiagnosticCollector;
import javax.tools.JavaFileObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Builds the SCIP index.
 *
 * <p>This class is responsible for building the SCIP index from a list of target roots, output
 * path, source root, and classpath. It uses the ScipSemanticdb class to build the index and the
 * UberScipSemanticdbReporter to report errors.
 */
public class ScipBuilder {

  private static final Logger logger = LoggerFactory.getLogger(ScipBuilder.class);
  private final ScipSemanticdbReporter reporter;
  private final ScipSemanticdbFactory scipSemanticdbFactory;
  private final DiagnosticCollector<JavaFileObject> diagnosticCollector;

  public ScipBuilder(DiagnosticCollector<JavaFileObject> diagnosticCollector) {
    this(new UberScipSemanticdbReporter(), new DefaultScipSemanticdbFactory(), diagnosticCollector);
  }

  public ScipBuilder(
      ScipSemanticdbReporter reporter,
      ScipSemanticdbFactory scipSemanticdbFactory,
      DiagnosticCollector<JavaFileObject> diagnosticCollector) {
    this.reporter = reporter;
    this.scipSemanticdbFactory = scipSemanticdbFactory;
    this.diagnosticCollector = diagnosticCollector;
  }

  /**
   * Builds the SCIP index.
   *
   * @param buildOptions BuildOptions containing target roots, output path, source root, and
   *     classpath
   */
  public void buildScip(BuildOptions buildOptions) throws IOException {

    List<MavenPackage> mavenPackages =
        collectMavenPackagesFromClasspath(buildOptions.getClasspathString());
    // Reverse it make sure that map will have the latest version of the package
    // @see com.sourcegraph.scip_semanticdb.PackageTable
    Collections.reverse(mavenPackages);

    ScipSemanticdbOptions scipOptions =
        new ScipSemanticdbOptions(
            buildOptions.getTargetRoots(),
            buildOptions.getOutputPath(),
            buildOptions.getSourceRoot(),
            this.reporter,
            LsifToolInfo.newBuilder().setName("scip-java").setVersion("HEAD").build(),
            "java",
            ScipOutputFormat.TYPED_PROTOBUF,
            true, // parallel,
            mavenPackages,
            /* buildKind */ "",
            /* emitInverseRelationships */ true,
            /* allowEmptyIndex */ true,
            /* indexDirectoryEntries */ true);
    ScipSemanticdb semanticdb =
        scipSemanticdbFactory.create(
            new UberScipWriter(scipOptions, buildOptions, this.diagnosticCollector), scipOptions);
    try {
      // Hate this, but there not much we can do.
      // We either have to use reflection or read index into memory after it gets written by
      // original writer.
      // In the long run we can create PR upstream. But for now, this is the best we can do.
      Method runMethod = ScipSemanticdb.class.getDeclaredMethod("run");
      runMethod.setAccessible(true);
      runMethod.invoke(semanticdb);
    } catch (NoSuchMethodException | IllegalAccessException | InvocationTargetException e) {
      logger.error("Failed to invoke run method via reflection", e);
      throw new RuntimeException("Failed to build SCIP index", e);
    }

    if (scipOptions.reporter.hasErrors()) {
      logger.debug("SCIP index generation failed");
    }
  }

  /**
   * Collects Maven package information from the provided classpath.
   *
   * @param classpath containing classpath information
   * @return List of MavenPackage objects extracted from the classpath
   */
  public List<MavenPackage> collectMavenPackagesFromClasspath(String classpath) {
    List<MavenPackage> mavenPackages = new ArrayList<>();

    try {
      // Get classpath from options
      if (classpath == null || classpath.isEmpty()) {
        logger.warn("Warning: Empty classpath provided");
        return mavenPackages;
      }

      // Split classpath by path separator (: on Unix, ; on Windows)
      String[] classpathEntries = classpath.split(System.getProperty("path.separator"));

      for (String entry : classpathEntries) {
        if (!is3rdPartyOutputPath(entry)) {
          continue;
        }

        Path path = Path.of(entry);

        // Check if the entry is a JAR file
        if (Files.exists(path) && entry.toLowerCase().endsWith(".jar")) {
          // Extract Maven coordinates from JAR filename or manifest
          MavenPackage mavenPackage = extractMavenInfoFromJar(path);
          mavenPackages.add(mavenPackage);
        }
      }
    } catch (Exception e) {
      logger.error("Error collecting Maven packages from classpath: " + e.getMessage(), e);
    }

    return mavenPackages;
  }

  /**
   * Extracts Maven package information from a JAR file.
   *
   * @param jarPath Path to the JAR file
   * @return MavenPackage object if Maven information can be extracted, null otherwise
   */
  public MavenPackage extractMavenInfoFromJar(Path jarPath) {
    // name will be used as version since we can't decode maven coordinates from jar
    String filename = jarPath.getFileName().toString().replace(".jar", "");
    return new MavenPackage(jarPath, ".", ".", filename);
  }
}
