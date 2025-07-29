package com.uber.scip.aggregator;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.util.Arrays;
import org.apache.commons.cli.ParseException;
import org.junit.jupiter.api.Test;

public class CommandLineConfigTest {

  @Test
  public void testParseArgs_MandatoryConfigOnly() throws ParseException {
    String[] args = {"-m", "config.properties"};
    CommandLineConfig config = CommandLineConfig.parseArgs(args);

    assertEquals("config.properties", config.configFile);
    assertNull(config.outputPath);
    assertNull(config.files);
    assertNull(config.rootDir);
  }

  @Test
  public void testParseArgs_AllOptions() throws ParseException {
    String[] args = {
      "-m", "config.properties",
      "-o", "output.scip",
      "-f", "file1.java,file2.java",
      "-r", "/path/to/root"
    };
    CommandLineConfig config = CommandLineConfig.parseArgs(args);

    assertEquals("config.properties", config.configFile);
    assertEquals("output.scip", config.outputPath);
    assertEquals(Arrays.asList("file1.java", "file2.java"), config.files);
    assertEquals("/path/to/root", config.rootDir);
  }

  @Test
  public void testParseArgs_MissingMandatoryConfig() {
    assertThrows(
        ParseException.class,
        () -> {
          String[] args = {"-o", "output.scip"};
          CommandLineConfig.parseArgs(args);
        });
  }

  @Test
  public void testParseArgs_WithOutput() throws ParseException {
    String[] args = {"-m", "config.properties", "-o", "output.scip"};
    CommandLineConfig config = CommandLineConfig.parseArgs(args);

    assertEquals("config.properties", config.configFile);
    assertEquals("output.scip", config.outputPath);
    assertNull(config.files);
    assertNull(config.rootDir);
  }

  @Test
  public void testParseArgs_WithFiles() throws ParseException {
    String[] args = {"-m", "config.properties", "-f", "file1.java,file2.java"};
    CommandLineConfig config = CommandLineConfig.parseArgs(args);

    assertEquals("config.properties", config.configFile);
    assertNull(config.outputPath);
    assertEquals(Arrays.asList("file1.java", "file2.java"), config.files);
    assertNull(config.rootDir);
  }

  @Test
  public void testParseArgs_WithRootDir() throws ParseException {
    String[] args = {"-m", "config.properties", "-r", "/path/to/root"};
    CommandLineConfig config = CommandLineConfig.parseArgs(args);

    assertEquals("config.properties", config.configFile);
    assertNull(config.outputPath);
    assertNull(config.files);
    assertEquals("/path/to/root", config.rootDir);
  }
}
