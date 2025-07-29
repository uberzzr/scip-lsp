package com.uber.scip.aggregator;

import java.nio.file.Paths;

/** Manages SemanticDB source and target roots for the Java analyzer. */
public class SemanticDbManager {
  private String sourceRoot;
  private String targetRoot;
  private String plugin;

  public SemanticDbManager() {
    this.sourceRoot = Paths.get("").toAbsolutePath().toString();
    this.targetRoot = Paths.get("").toAbsolutePath().toString();
    this.plugin = Paths.get("").toAbsolutePath().toString();
  }

  /** Sets the SemanticDB source root. */
  public void setSourceRoot(String sourceRoot) {
    this.sourceRoot = sourceRoot;
  }

  /** Sets the SemanticDB target root. */
  public void setTargetRoot(String targetRoot) {
    this.targetRoot = targetRoot;
  }

  /** Returns the SemanticDB source root. */
  public String getSourceRoot() {
    return sourceRoot;
  }

  /** Returns the SemanticDB target root. */
  public String getTargetRoot() {
    return targetRoot;
  }

  /** Checks if a SemanticDB source root has been defined. */
  public boolean hasSourceRoot() {
    return sourceRoot != null && !sourceRoot.isEmpty();
  }

  /** Checks if a SemanticDB target root has been defined. */
  public boolean hasTargetRoot() {
    return targetRoot != null && !targetRoot.isEmpty();
  }

  /** Returns the SemanticDB source root option string for the compiler. */
  public String getSourceRootOption() {
    return "-sourceroot:" + sourceRoot;
  }

  public String getPlugin() {
    return plugin;
  }

  public void setPlugin(String plugin) {
    this.plugin = plugin;
  }

  /** Returns the SemanticDB target root option string for the compiler. */
  public String getTargetRootOption() {
    return "-targetroot:" + targetRoot;
  }

  public String formatSemanticDBPluginConfig() {
    return "-Xplugin:semanticdb "
        + (this.hasSourceRoot() ? this.getSourceRootOption() : ".")
        + " "
        + (this.hasTargetRoot() ? this.getTargetRootOption() : "");
  }
}
