package com.uber.scip.aggregator;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Properties;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import org.jspecify.annotations.Nullable;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/** Loads configuration for the Aggregator from properties files. */
public class ConfigLoader {
  private static final Logger logger = LoggerFactory.getLogger(ConfigLoader.class);

  /** Loads configuration from a properties file. */
  public static FileAnalyzer loadFromFile(String configFilePath, @Nullable String rootDir)
      throws IOException {
    Properties properties = new Properties();
    try (FileInputStream fis = new FileInputStream(configFilePath)) {
      properties.load(fis);
    }

    FileAnalyzer aggregator = new FileAnalyzer();

    // Load classpath
    String classpath = properties.getProperty("classpath");
    if (classpath != null && !classpath.isEmpty()) {
      // Qualify classpath with rootDir if it's a relative path
      String qualifiedClasspath =
          String.join(":", qualifyPaths(Arrays.asList(classpath.split(":")), rootDir));
      aggregator.withClasspath(qualifiedClasspath);
    }

    // Load compiler options
    String options = properties.getProperty("options");
    if (options != null && !options.isEmpty()) {
      aggregator.withOptions(Arrays.asList(options.split(",")));
    }

    // Load SemanticDB source root
    String sourceRoot = properties.getProperty("semanticdb.sourceRoot");
    if (sourceRoot != null && !sourceRoot.isEmpty()) {
      // Qualify source root with rootDir if it's a relative path
      String qualifiedSourceRoot = qualifyPath(sourceRoot.trim(), rootDir);
      aggregator.withSemanticDbSourceRoot(qualifiedSourceRoot);
    }

    // Load SemanticDB target root
    String targetRoot = properties.getProperty("semanticdb.targetRoot");
    if (targetRoot != null && !targetRoot.isEmpty()) {
      // Qualify target root with rootDir if it's a relative path
      String qualifiedTargetRoot = qualifyPath(targetRoot, rootDir);
      aggregator.withSemanticDbTargetRoot(qualifiedTargetRoot);
    }

    // Load SemanticDB target root
    String semanticdbPlugin = properties.getProperty("semanticdb_plugin");
    if (semanticdbPlugin != null && !semanticdbPlugin.isEmpty()) {
      // Qualify semanticdb plugin path with rootDir if it's a relative path
      String qualifiedSemanticdbPlugin = qualifyPath(semanticdbPlugin, rootDir);
      aggregator.withSemanticDbPlugin(qualifiedSemanticdbPlugin);
    }

    // Load files to analyze
    String files = properties.getProperty("files");
    if (files != null && !files.isEmpty()) {
      List<String> filesList = Arrays.asList(files.split(","));
      // Qualify file paths with rootDir if they're relative paths
      List<String> qualifiedFiles = qualifyPaths(filesList, rootDir);
      aggregator.withFiles(qualifiedFiles);
    }
    // Ability to load files from a file
    String listOfFiles = properties.getProperty("files_file");
    if (listOfFiles != null && !listOfFiles.isEmpty()) {

      // Qualify the files_file path with rootDir if it's a relative path
      List<String> paths = Files.readAllLines(Paths.get(qualifyPath(listOfFiles, rootDir)));
      // Qualify file paths with rootDir if they're relative paths
      List<String> qualifiedPaths = qualifyPaths(paths, rootDir);
      aggregator.withFiles(qualifiedPaths);
    }

    // Load files to analyze
    String output = properties.getProperty("output");
    if (output != null && !output.isEmpty()) {
      // Qualify output path with rootDir if it's a relative path
      String qualifiedOutput = qualifyPath(output, rootDir);
      aggregator.withOutputPath(qualifiedOutput);
    }

    return aggregator;
  }

  /** Finds Java files from a list of paths. */
  public static List<File> findJavaFilesFromPaths(List<String> paths) {
    List<File> javaFiles = new ArrayList<>();

    for (String path : paths) {
      File file = new File(path.trim());
      if (file.isFile() && path.endsWith(".java")) {
        javaFiles.add(file);
      } else if (file.isDirectory()) {
        try (Stream<Path> stream = Files.walk(Paths.get(path))) {
          List<File> filesInDir =
              stream
                  .filter(p -> p.toString().endsWith(".java"))
                  .map(Path::toFile)
                  .collect(Collectors.toList());
          javaFiles.addAll(filesInDir);
        } catch (IOException e) {
          logger.warn("Error walking directory {}: {}", path, e.getMessage());
        }
      }
    }

    return javaFiles;
  }

  private static String qualifyPath(String path, @Nullable String rootDir) {
    if (rootDir == null || rootDir.isEmpty() || path.startsWith("/")) {
      return path;
    }
    return Paths.get(rootDir, path).toString();
  }

  /**
   * Qualifies a list of file paths with the root directory if they're relative paths.
   *
   * @param paths The list of paths to qualify
   * @param rootDir The root directory to qualify the paths with
   * @return The qualified file paths
   */
  private static List<String> qualifyPaths(List<String> paths, @Nullable String rootDir) {
    if (rootDir == null || rootDir.isEmpty()) {
      return paths;
    }
    return paths.stream().map(path -> qualifyPath(path, rootDir)).collect(Collectors.toList());
  }
}
