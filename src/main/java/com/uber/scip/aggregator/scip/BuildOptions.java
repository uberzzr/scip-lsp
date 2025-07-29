package com.uber.scip.aggregator.scip;

import java.nio.file.Path;
import java.util.List;

public class BuildOptions {
  private final List<Path> targetRoots;
  private final Path sourceRoot;
  private final Path outputPath;
  private final String classpathString;
  private final boolean shadeIdlSymbols;
  private final boolean shade3rdPartySymbols;

  private BuildOptions(Builder builder) {
    this.targetRoots = builder.targetRoots;
    this.sourceRoot = builder.sourceRoot;
    this.outputPath = builder.outputPath;
    this.classpathString = builder.classpathString;
    this.shadeIdlSymbols = builder.shadeIdlSymbols;
    this.shade3rdPartySymbols = builder.shade3rdPartySymbols;
  }

  public List<Path> getTargetRoots() {
    return targetRoots;
  }

  public Path getSourceRoot() {
    return sourceRoot;
  }

  public Path getOutputPath() {
    return outputPath;
  }

  public String getClasspathString() {
    return classpathString;
  }

  public boolean shouldShadeIdlSymbols() {
    return shadeIdlSymbols;
  }

  public boolean shouldShade3rdPartySymbols() {
    return shade3rdPartySymbols;
  }

  public static BuildOptions defaultOptions() {
    return new Builder().build();
  }

  public static class Builder {
    private List<Path> targetRoots = List.of();
    private Path sourceRoot = Path.of("");
    private Path outputPath = Path.of("");
    private String classpathString = "";
    private boolean shadeIdlSymbols = true;
    private boolean shade3rdPartySymbols = false;

    public Builder withTargetRoots(List<Path> targetRoots) {
      this.targetRoots = targetRoots;
      return this;
    }

    public Builder withSourceRoot(Path sourceRoot) {
      this.sourceRoot = sourceRoot;
      return this;
    }

    public Builder withOutputPath(Path outputPath) {
      this.outputPath = outputPath;
      return this;
    }

    public Builder withClasspathString(String classpathString) {
      this.classpathString = classpathString;
      return this;
    }

    public Builder shouldShadeIdlSymbols(boolean shadeIdlSymbols) {
      this.shadeIdlSymbols = shadeIdlSymbols;
      return this;
    }

    public Builder shouldShade3rdPartySymbols(boolean shade3rdPartySymbols) {
      this.shade3rdPartySymbols = shade3rdPartySymbols;
      return this;
    }

    public BuildOptions build() {
      return new BuildOptions(this);
    }
  }
}
