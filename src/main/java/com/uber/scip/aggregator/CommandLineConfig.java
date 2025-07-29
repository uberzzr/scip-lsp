package com.uber.scip.aggregator;

import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import org.apache.commons.cli.CommandLine;
import org.apache.commons.cli.CommandLineParser;
import org.apache.commons.cli.DefaultParser;
import org.apache.commons.cli.HelpFormatter;
import org.apache.commons.cli.Options;
import org.apache.commons.cli.ParseException;
import org.jspecify.annotations.Nullable;

public class CommandLineConfig {
  final String configFile;
  @Nullable final String outputPath;
  @Nullable final List<String> files;
  @Nullable final String rootDir;

  CommandLineConfig(
      String configFile,
      @Nullable String outputPath,
      @Nullable List<String> files,
      @Nullable String rootDir) {
    this.configFile = configFile;
    this.outputPath = outputPath;
    this.files = files;
    this.rootDir = rootDir;
  }

  static CommandLineConfig parseArgs(String[] args) throws ParseException {
    Options options = createCommandLineOptions();
    CommandLineParser parser = new DefaultParser();
    CommandLine cmd = parser.parse(options, args);

    String configFile = cmd.getOptionValue("manifest-configuration");
    String outputPath = cmd.getOptionValue("output");
    List<String> files =
        Optional.ofNullable(cmd.getOptionValue("files"))
            .map(o -> Arrays.asList(o.split(",")))
            .orElse(null);
    String rootDir = cmd.getOptionValue("root-dir");

    return new CommandLineConfig(configFile, outputPath, files, rootDir);
  }

  static void printHelp() {
    new HelpFormatter()
        .printHelp(
            "java Aggregator",
            "\nAnalyzes Java files and generates SCIP index.\n\n",
            createCommandLineOptions(),
            "\nExample: java Aggregator -m config.properties -f file1.java,file2.java -o output.scip\n",
            true);
  }

  private static Options createCommandLineOptions() {
    Options options = new Options();
    options.addRequiredOption(
        "m", "manifest-configuration", true, "Configuration file path (mandatory)");
    options.addOption("f", "files", true, "Comma-separated list of files to analyze (optional)");
    options.addOption("o", "output", true, "Output file path (optional)");
    options.addOption(
        "r", "root-dir", true, "Root directory to qualify files and classpath (optional)");
    return options;
  }
}
