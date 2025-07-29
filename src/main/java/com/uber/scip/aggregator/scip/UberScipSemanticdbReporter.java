package com.uber.scip.aggregator.scip;

import com.sourcegraph.scip_semanticdb.ScipSemanticdbReporter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/** A reporter for the SCIP semanticdb plugin. */
public class UberScipSemanticdbReporter extends ScipSemanticdbReporter {

  private static final Logger logger = LoggerFactory.getLogger(UberScipSemanticdbReporter.class);

  private boolean hasErrors = false;

  @Override
  public void error(Throwable e) {
    e.printStackTrace(System.err);
    hasErrors = true;
  }

  @Override
  public void error(String message) {
    logger.error("ERROR[scip-semanticdb]: {}", message);
    hasErrors = true;
  }

  @Override
  public boolean hasErrors() {
    return this.hasErrors;
  }
}
