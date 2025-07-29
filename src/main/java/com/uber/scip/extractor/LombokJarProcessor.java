package com.uber.scip.extractor;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.nio.file.attribute.FileTime;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.jar.JarEntry;
import java.util.jar.JarFile;
import java.util.jar.JarOutputStream;
import org.objectweb.asm.ClassReader;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class LombokJarProcessor {

  // This is a future time to set the last modified time of the JAR entries.
  // This is used to ensure that the generated classes are always considered newer than
  // the original classes, which is important for javac to pick them up. See "-Xprefer:newer".
  public static final Long FUTURE_TIME = 2000000000000L; // GMT: Wednesday, May 18, 2033 3:33:20 AM

  static {
    // Set these properties before any logging occurs
    // IMPORTANT: This must be the first code that runs in the class
    System.setProperty("org.slf4j.simpleLogger.showDateTime", "true");
    System.setProperty("org.slf4j.simpleLogger.dateTimeFormat", "yyyy-MM-dd HH:mm:ss.SSS");
    System.setProperty("org.slf4j.simpleLogger.defaultLogLevel", "INFO");
    System.setProperty("org.slf4j.simpleLogger.showLogName", "true");
    System.setProperty("org.slf4j.simpleLogger.showShortLogName", "true");
    System.setProperty("org.slf4j.simpleLogger.logFile", "System.out");
  }

  private static final Logger logger = LoggerFactory.getLogger(LombokJarProcessor.class);

  private File validateInputJar(String inputJarPath) {
    File inputJar = new File(inputJarPath);
    if (!inputJar.exists()) {
      throw new IllegalArgumentException("Input JAR file does not exist: " + inputJarPath);
    }
    if (!inputJar.isFile()) {
      throw new IllegalArgumentException("Input path is not a file: " + inputJarPath);
    }
    return inputJar;
  }

  private List<JarEntryInfo> findLombokClasses(File inputJar) throws IOException {
    List<JarEntryInfo> lombokClasses = new ArrayList<>();

    Set<String> lombokClassesSet = new HashSet<>();

    try (JarFile jarIn = new JarFile(inputJar)) {
      jarIn
          .entries()
          .asIterator()
          .forEachRemaining(
              entry -> {
                if (entry.getName().endsWith(".class")) {
                  try {
                    if (isLombokGeneratedClass(entry, jarIn)) {
                      String parentClassName = entry.getName();
                      if (entry.getName().contains("$")) {
                        parentClassName =
                            entry.getName().substring(0, entry.getName().indexOf("$"));
                      }
                      parentClassName = parentClassName.replace(".class", "");
                      lombokClassesSet.add(parentClassName);
                    }
                  } catch (IOException e) {
                    logger.warn("Failed to analyze class file: " + entry.getName(), e);
                  }
                }
              });
      jarIn
          .entries()
          .asIterator()
          .forEachRemaining(
              entry -> {
                if (entry.getName().endsWith(".class")) {
                  String parentClassName = entry.getName();
                  if (entry.getName().contains("$")) {
                    // lombok annotation is defined in a nested class
                    parentClassName = entry.getName().substring(0, entry.getName().indexOf("$"));
                  }
                  parentClassName = parentClassName.replace(".class", "");
                  try {
                    if (lombokClassesSet.contains(parentClassName)) {
                      lombokClasses.add(
                          new JarEntryInfo(
                              entry.getName(), jarIn.getInputStream(entry).readAllBytes()));
                      logger.debug("Found Lombok-generated class: {}", entry.getName());
                    }
                  } catch (IOException e) {
                    logger.warn("Failed to analyze class file: " + entry.getName(), e);
                  }
                }
              });
    }

    logger.debug("Found {} Lombok-generated classes", lombokClasses.size());
    return lombokClasses;
  }

  private void createOutputJar(String outputJarPath, List<JarEntryInfo> lombokClasses)
      throws IOException {
    try (JarOutputStream jarOut = new JarOutputStream(new FileOutputStream(outputJarPath))) {
      for (JarEntryInfo entryInfo : lombokClasses) {
        JarEntry newEntry = new JarEntry(entryInfo.name());
        newEntry.setLastModifiedTime(FileTime.fromMillis(FUTURE_TIME));
        jarOut.putNextEntry(newEntry);
        jarOut.write(entryInfo.content());
        jarOut.closeEntry();
      }
      new File(outputJarPath).setLastModified(FUTURE_TIME);
    }
  }

  private boolean isLombokGeneratedClass(JarEntry entry, JarFile jarIn) throws IOException {
    if (!entry.getName().endsWith(".class")) {
      return false;
    }

    try {
      ClassReader reader = new ClassReader(jarIn.getInputStream(entry));
      LombokClassVisitor classVisitor = new LombokClassVisitor();
      reader.accept(classVisitor, ClassReader.SKIP_DEBUG | ClassReader.SKIP_FRAMES);
      return classVisitor.isLombokGenerated();
    } catch (IOException e) {
      logger.warn("Failed to analyze class file: " + entry.getName(), e);
      return false;
    }
  }

  public void processJar(String inputJarPath, String outputJarPath) throws IOException {
    File inputJar = validateInputJar(inputJarPath);

    // If the file is provided to javac as explicit source - lombok will work,
    // but if the file will be found by classloader (i.e. as a dependency)
    // lombok will not be processed for that file, thus we need to shortcut
    // class lookup and point it to the already generated file.
    // -Xprefer:newer will enforce that. If the lombok POJO is inner class, we have to cache
    // the parent class also, otherwise javac will find it, and will process it with errors.
    List<JarEntryInfo> lombokClasses = findLombokClasses(inputJar);
    createOutputJar(outputJarPath, lombokClasses);

    logger.debug("Successfully processed JAR file. Output: {}", outputJarPath);
  }

  public static void main(String[] args) {
    if (args.length < 2) {
      System.exit(1);
    }

    String inputJarPath = args[0];
    String outputJarPath = args[1];

    try {
      new LombokJarProcessor().processJar(inputJarPath, outputJarPath);
    } catch (Exception e) {
      logger.error("Failed to process JAR file", e);
      System.exit(1);
    }
  }

  private record JarEntryInfo(String name, byte[] content) {}
}
