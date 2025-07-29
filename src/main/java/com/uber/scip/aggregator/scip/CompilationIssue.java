package com.uber.scip.aggregator.scip;

import java.io.IOException;
import javax.tools.Diagnostic;
import javax.tools.JavaFileObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/** Represents a compilation error or warning */
public class CompilationIssue {
  private static final Logger logger = LoggerFactory.getLogger(CompilationIssue.class);

  private final Diagnostic.Kind kind;
  private final String message;
  private final String source;
  private final long lineNumber;
  private final long columnNumberStart;
  private final long columnNumberEnd;

  public CompilationIssue(Diagnostic<? extends JavaFileObject> diagnostic) {
    this.kind = diagnostic.getKind();
    this.message = diagnostic.getMessage(null);
    this.source = diagnostic.getSource() != null ? diagnostic.getSource().getName() : "unknown";
    this.lineNumber =
        diagnostic.getSource() != null
            ? diagnostic.getLineNumber() - 1
            : 0; // Convert to 0-based index

    // Get the source content if available
    String sourceContent = null;
    if (diagnostic.getSource() != null) {
      try {
        sourceContent = diagnostic.getSource().getCharContent(true).toString();
      } catch (IOException e) {
        logger.warn(e.getMessage());
      }
    }

    // Calculate column positions based on source content,
    // this might be suboptimal since for every issue we read the source file
    if (sourceContent != null) {
      String[] lines = sourceContent.split("\n", -1);
      if (lineNumber >= 0 && lineNumber <= lines.length) {
        long startPos = diagnostic.getStartPosition();
        long endPos = diagnostic.getEndPosition();

        // Count characters up to this line
        long previousLinesChars = 0;
        for (int i = 0; i < lineNumber - 1; i++) {
          previousLinesChars += lines[i].length() + 1; // +1 for newline
        }

        // Calculate column positions within the line
        this.columnNumberStart = startPos - previousLinesChars - 1;
        this.columnNumberEnd = endPos - previousLinesChars;
      } else {
        this.columnNumberStart = 0;
        this.columnNumberEnd = 0;
      }
    } else {
      // Fallback to basic column number if source content not available
      this.columnNumberStart = diagnostic.getColumnNumber() >= 0 ? diagnostic.getColumnNumber() : 0;
      this.columnNumberEnd = this.columnNumberStart;
    }
  }

  public Diagnostic.Kind getKind() {
    return kind;
  }

  public String getMessage() {
    return message;
  }

  public String getSource() {
    return source;
  }

  public long getLineNumber() {
    return lineNumber;
  }

  public long getColumnNumberStart() {
    return columnNumberStart;
  }

  public long getColumnNumberEnd() {
    return columnNumberEnd;
  }

  @Override
  public String toString() {
    return kind
        + " in "
        + source
        + " at line "
        + lineNumber
        + ", column "
        + columnNumberStart
        + ": "
        + columnNumberEnd
        + ": "
        + message;
  }
}
