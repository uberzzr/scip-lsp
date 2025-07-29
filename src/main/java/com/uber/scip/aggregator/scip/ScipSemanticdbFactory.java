package com.uber.scip.aggregator.scip;

import com.sourcegraph.scip_semanticdb.ScipSemanticdb;
import com.sourcegraph.scip_semanticdb.ScipSemanticdbOptions;
import com.sourcegraph.scip_semanticdb.ScipWriter;

// New interface for factory
public interface ScipSemanticdbFactory {
  ScipSemanticdb create(ScipWriter writer, ScipSemanticdbOptions options);
}
