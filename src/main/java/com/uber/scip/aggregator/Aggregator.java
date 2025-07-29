package com.uber.scip.aggregator;

import java.io.File;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import org.apache.commons.cli.ParseException;
import org.jspecify.annotations.Nullable;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/** A tool that uses the JavaC Compiler API to analyze Java files. */
public class Aggregator {
  static {
    // Set these properties before any logging occurs
    // IMPORTANT: This must be the first code that runs in the class
    System.setProperty("org.slf4j.simpleLogger.showDateTime", "true");
    System.setProperty("org.slf4j.simpleLogger.dateTimeFormat", "yyyy-MM-dd HH:mm:ss.SSS");
    System.setProperty("org.slf4j.simpleLogger.defaultLogLevel", "INFO");
    System.setProperty("org.slf4j.simpleLogger.showLogName", "true");
    System.setProperty("org.slf4j.simpleLogger.showShortLogName", "true");
    System.setProperty("org.slf4j.simpleLogger.logFile", "System.out");
  }

  private static final Logger logger = LoggerFactory.getLogger(Aggregator.class);

  /** Loads configuration from a properties file */
  public static FileAnalyzer fromConfigFile(String configFilePath, @Nullable String rootDir)
      throws IOException {
    return ConfigLoader.loadFromFile(configFilePath, rootDir);
  }

  public static void main(String[] args) {
    try {
      CommandLineConfig config = CommandLineConfig.parseArgs(args);

      // Load analyzer from config file
      FileAnalyzer analyzer = fromConfigFile(config.configFile, config.rootDir);
      logger.debug("Loaded configuration from: {}", config.configFile);

      // Apply command line overrides
      if (config.outputPath != null) {
        analyzer.withOutputPath(config.outputPath);
        logger.debug("Overriding output path to: {}", config.outputPath);
      }

      // Process files
      List<File> javaFiles = new ArrayList<>();
      if (config.files != null) {
        analyzer.withFiles(config.files);
        javaFiles = ConfigLoader.findJavaFilesFromPaths(config.files);
        logger.debug("Using files from command line arguments: {}", config.files);
      } else if (!analyzer.getFilePaths().isEmpty()) {
        javaFiles = ConfigLoader.findJavaFilesFromPaths(analyzer.getFilePaths());
        logger.debug("Using files specified in config: {}", analyzer.getFilePaths());
      }

      if (javaFiles.isEmpty()) {
        logger.debug("No Java files found in the specified paths or source roots");
      }

      // Log configuration
      logger.debug("Found {} Java files", javaFiles.size());
      logger.debug("SemanticDB Source Root: {}", analyzer.getSemanticDbManager().getSourceRoot());
      logger.debug("SemanticDB Target Root: {}", analyzer.getSemanticDbManager().getTargetRoot());
      logger.debug("Classpath: {}", analyzer.getCompilerOptions().getClasspath());
      logger.debug("Options: {}", analyzer.getCompilerOptions().getOptions());
      logger.debug("Output: {}", analyzer.getOutputPath());

      // Perform analysis
      analyzer.analyzeFiles(javaFiles);
      logger.debug("Analysis complete");

    } catch (ParseException | IOException e) {
      String message = String.format("Error parsing command line arguments: %s", e.getMessage());
      logger.debug(message);
      CommandLineConfig.printHelp();
      throw new IllegalArgumentException(message);
    }
  }
}
