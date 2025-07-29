package com.uber.intellij.jd;

import static com.uber.intellij.jd.Decompiler.DECOMPILER_OUTPUT_FILE_NAME_KEY_PREFIX;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.BufferedReader;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.jar.JarFile;

import com.uber.utils.TestUtils;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

/** Unit tests for the {@link Decompiler}. */
public class DecompilerTest {

    private Path tmpDir;

    private static final Path RESOURCE_JAR = TestUtils
            .getResourcePath("src/test/java/com/uber/intellij/jd/resources/libsrc_main.jar");

    @TempDir public File tempFolder;

    @BeforeEach
    public void setUp() throws IOException {
        tmpDir = Files.createTempDirectory("tmp");
    }

    @Test
    public void testDecompiler_decompilesJar() throws IOException {
        String expectedFile = tmpDir.resolve("libsrc_main_custom_name.jar").toString();
        String[] args =
                new String[] {
                        DECOMPILER_OUTPUT_FILE_NAME_KEY_PREFIX + expectedFile, RESOURCE_JAR.toString(),
                };

        assertFalse(new File(expectedFile).exists());
        Decompiler.main(args);
        assertTrue(new File(expectedFile).exists());
        // check file creation time
        assertEquals(
                0,
                Files.getLastModifiedTime(Paths.get(expectedFile)).toMillis(),
                "Creation time should be 0");

        String jarPath = getJarPath();
        if (jarPath == null || jarPath.length() == 0) {
            throw new RuntimeException("JARBIN_PATH env variable should be set");
        }

        ProcessBuilder pb = new ProcessBuilder(jarPath, "tf", expectedFile);
        Process process = pb.start();
        BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()));
        StringBuilder builder = new StringBuilder();
        String line = null;
        while ((line = reader.readLine()) != null) {
            builder.append(line);
            builder.append(System.getProperty("line.separator"));
        }
        String result = builder.toString();
        assertTrue(result.contains("Decompiler.java"));
    }

    @Test
    public void testDecompiler_decompilesJar_wrongOutputPath() {
        assertThrows(
                IOException.class,
                () -> {
                    String expectedFile = "/dev/null/libsrc_main_custom_name.jar";
                    String[] args =
                            new String[] {
                                    DECOMPILER_OUTPUT_FILE_NAME_KEY_PREFIX + expectedFile, RESOURCE_JAR.toString(),
                            };

                    assertFalse(new File(expectedFile).exists());
                    Decompiler.main(args);
                });
    }

    // SecurityException since System.exit is called in test.
    @Test
    public void testDecompiler_throws_on_timeout() {
        assertThrows(
                SecurityException.class,
                () -> {
                    String[] args = new String[] {"-timeout=-1", RESOURCE_JAR.toString(), tmpDir.toString()};

                    Decompiler.main(args);
                });
    }

    // SecurityException since System.exit is called in test.
    @Test
    public void testDecompiler_throws_incomplete_args() {
        assertThrows(
                SecurityException.class,
                () -> {
                    String[] args = new String[] {"-timeout=1", "-max_files=10"};

                    Decompiler.main(args);
                });
    }

    // SecurityException since System.exit is called in test.
    @Test
    public void testDecompiler_throws_on_missing_args() {
        assertThrows(
                SecurityException.class,
                () -> {
                    String[] args = new String[] {"-timeout=-1"};

                    Decompiler.main(args);
                });
    }

    @Test
    public void testCountFilesInJar_withValidJar() throws IOException {
        // Test with the existing test JAR
        long count = Decompiler.countFilesInJar(RESOURCE_JAR.toString());
        assertTrue(count > 0);

        String jarPath = getJarPath();
        if (jarPath == null || jarPath.isEmpty()) {
            throw new RuntimeException("JARBIN_PATH env variable should be set");
        }
        int manualFileCount = getManualFileCount(jarPath);

        assertEquals(manualFileCount, count);
    }

    /**
     * Test that the main function creates an empty JAR and skips decompilation when the JAR size is
     * greater than the specified limit.
     */
    @Test
    public void testMain_createsEmptyJarWhenSizeExceedsLimit() throws IOException {
        int actualFileCount = getManualFileCount(RESOURCE_JAR.toString());
        String maxFilesArg = "-max_files=" + (actualFileCount - 1);

        ByteArrayOutputStream errContent = new ByteArrayOutputStream();
        PrintStream originalErr = System.err;
        System.setErr(new PrintStream(errContent));

        File outputDir = newFolder(tempFolder, "output");
        String inputJarName = RESOURCE_JAR.getFileName().toString();
        String expectedOutputJarPath = new File(outputDir, inputJarName).getAbsolutePath();

        File outputJarFile = new File(expectedOutputJarPath);
        if (outputJarFile.exists()) {
            outputJarFile.delete();
        }

        String[] args =
                new String[] {
                        DECOMPILER_OUTPUT_FILE_NAME_KEY_PREFIX + expectedOutputJarPath,
                        maxFilesArg,
                        RESOURCE_JAR.toString(),
                };
        Decompiler.main(args);

        System.setErr(originalErr);

        assertTrue(new File(expectedOutputJarPath).exists());
        assertTrue(isValidJar(expectedOutputJarPath));
        String errOutput = errContent.toString();
        assertTrue(errOutput.contains("More than allowed threshold"));
        assertTrue(errOutput.contains("Creating empty JAR"));
    }

    /** Check if a file is a valid JAR by trying to open it with JarFile. */
    private boolean isValidJar(String jarPath) {
        try (JarFile jarFile = new JarFile(jarPath)) {
            return true;
        } catch (IOException e) {
            return false;
        }
    }

    private static int getManualFileCount(String jarPath) throws IOException {
        ProcessBuilder pb = new ProcessBuilder(jarPath, "tf", RESOURCE_JAR.toString());
        Process process = pb.start();
        BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()));

        int manualFileCount = 0;
        String line;

        while ((line = reader.readLine()) != null) {
            if (!line.trim().isEmpty() && line.endsWith(".class")) {
                manualFileCount++;
            }
        }
        return manualFileCount;
    }

    private static File newFolder(File root, String... subDirs) throws IOException {
        String subFolder = String.join("/", subDirs);
        File result = new File(root, subFolder);
        if (!result.mkdirs()) {
            throw new IOException("Couldn't create folders " + root);
        }
        return result;
    }

    private static String getJarPath() {
        String runfilesDir = System.getenv("TEST_SRCDIR");
        String jarPath = System.getenv("JARBIN_PATH").substring("external/".length());
        return Path.of(runfilesDir, jarPath).toString();
    }
}

