package com.uber.scip.aggregator.scip;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.anyBoolean;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.io.File;
import java.io.IOException;
import javax.tools.Diagnostic;
import javax.tools.JavaFileObject;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

public class CompilationIssuesTest {

  @TempDir public File tempFolder;

  @Test
  public void testCompilationIssueCreation() throws IOException {
    // Mock a diagnostic
    @SuppressWarnings("unchecked")
    Diagnostic<JavaFileObject> diagnostic = mock(Diagnostic.class);
    JavaFileObject source = mock(JavaFileObject.class);

    when(diagnostic.getKind()).thenReturn(Diagnostic.Kind.ERROR);
    when(diagnostic.getMessage(null)).thenReturn("Test error message");
    when(diagnostic.getSource()).thenReturn(source);
    when(source.getName()).thenReturn("TestFile.java");
    when(source.getCharContent(anyBoolean())).thenReturn("Java file content");
    when(diagnostic.getLineNumber()).thenReturn(1L);
    when(diagnostic.getColumnNumber()).thenReturn(5L);
    when(diagnostic.getStartPosition()).thenReturn(2L);
    when(diagnostic.getEndPosition()).thenReturn(5L);

    CompilationIssue issue = new CompilationIssue(diagnostic);

    assertEquals(Diagnostic.Kind.ERROR, issue.getKind());
    assertEquals("Test error message", issue.getMessage());
    assertEquals("TestFile.java", issue.getSource());
    assertEquals(0L, issue.getLineNumber());
    assertEquals(1L, issue.getColumnNumberStart());
    assertEquals(5L, issue.getColumnNumberEnd());
  }

  @Test
  public void testCompilationIssueWithNullSource() {
    // Mock a diagnostic with null source
    @SuppressWarnings("unchecked")
    Diagnostic<JavaFileObject> diagnostic = mock(Diagnostic.class);

    when(diagnostic.getKind()).thenReturn(Diagnostic.Kind.WARNING);
    when(diagnostic.getMessage(null)).thenReturn("Test warning message");
    when(diagnostic.getSource()).thenReturn(null);
    when(diagnostic.getLineNumber()).thenReturn(15L);
    when(diagnostic.getColumnNumber()).thenReturn(5L);

    CompilationIssue issue = new CompilationIssue(diagnostic);

    assertEquals(Diagnostic.Kind.WARNING, issue.getKind());
    assertEquals("Test warning message", issue.getMessage());
    assertEquals("unknown", issue.getSource());
    assertEquals(0L, issue.getLineNumber()); // no Surce, so line number is 0
    assertEquals(5L, issue.getColumnNumberStart());
  }

  @Test
  public void testToString() throws IOException {
    // Mock a diagnostic
    @SuppressWarnings("unchecked")
    Diagnostic<JavaFileObject> diagnostic = mock(Diagnostic.class);
    JavaFileObject source = mock(JavaFileObject.class);

    when(diagnostic.getKind()).thenReturn(Diagnostic.Kind.ERROR);
    when(diagnostic.getMessage(null)).thenReturn("Missing semicolon");
    when(source.getCharContent(anyBoolean())).thenReturn("Java file content");
    when(diagnostic.getSource()).thenReturn(source);
    when(source.getName()).thenReturn("Example.java");
    when(diagnostic.getLineNumber()).thenReturn(20L);
    when(diagnostic.getColumnNumber()).thenReturn(30L);

    CompilationIssue issue = new CompilationIssue(diagnostic);

    String expected = "ERROR in Example.java at line 19, column 0: 0: Missing semicolon";
    assertEquals(expected, issue.toString());
  }
}
