package com.uber.scip.aggregator.scip;

import com.sourcegraph.scip_semanticdb.ScipSemanticdb;
import com.sourcegraph.scip_semanticdb.ScipSemanticdbOptions;
import com.sourcegraph.scip_semanticdb.ScipWriter;

// Default implementation
public class DefaultScipSemanticdbFactory implements ScipSemanticdbFactory {
  @Override
  public ScipSemanticdb create(ScipWriter writer, ScipSemanticdbOptions options) {
    return new ScipSemanticdb(writer, options);
  }
}
