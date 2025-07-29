package com.uber.scip.aggregator;

import java.util.ArrayList;
import java.util.List;

/** Manages compiler options for the JavaC API */
public class CompilerOptions {
  private final List<String> options;
  private String classpath;

  public CompilerOptions() {
    this.options = new ArrayList<>();
    this.classpath = "";

    // Add default JDK module exports for compiler API access
    addJdkModuleExports();
  }

  private void addJdkModuleExports() {
    // Add options to make the compiler API accessible
    options.add("--add-exports");
    options.add("jdk.compiler/com.sun.tools.javac.api=ALL-UNNAMED");
    options.add("--add-exports");
    options.add("jdk.compiler/com.sun.tools.javac.util=ALL-UNNAMED");
    options.add("--add-exports");
    options.add("jdk.compiler/com.sun.tools.javac.tree=ALL-UNNAMED");
    options.add("--add-exports");
    options.add("jdk.compiler/com.sun.tools.javac.code=ALL-UNNAMED");
    // Stop even earlier, this still gets the result we need. FLOW is not working with -parameters
    options.add("-XDshould-stop.at=ATTR");
    // Enforce prefference to use newer versions of classes, we will be able to stop unnecessary
    // compilation with this.
    options.add("-Xprefer:newer");
  }

  public void addOptions(List<String> newOptions) {
    this.options.addAll(newOptions);
  }

  public void setClasspath(String classpath) {
    this.classpath = classpath;
  }

  public List<String> getOptions() {
    return options;
  }

  public String getClasspath() {
    return classpath;
  }

  public List<String> getCompilerOptions() {
    List<String> allOptions = new ArrayList<>(options);

    // Add classpath if specified
    if (!classpath.isEmpty()) {
      allOptions.add("-classpath");
      allOptions.add(classpath);
    }

    return allOptions;
  }
}
