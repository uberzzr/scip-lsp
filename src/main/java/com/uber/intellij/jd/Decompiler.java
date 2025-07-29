package com.uber.intellij.jd;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.attribute.BasicFileAttributeView;
import java.nio.file.attribute.FileTime;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;
import java.util.jar.JarFile;
import java.util.jar.JarOutputStream;
import java.util.stream.Stream;
import java.util.zip.ZipEntry;
import org.jetbrains.java.decompiler.main.decompiler.ConsoleDecompiler;

/**
 * The Decompiler class provides a main method to decompile Java classes from a JAR file and output
 * the decompiled source code to a specified directory.
 */
public class Decompiler {

    public static final String DECOMPILER_TIMEOUT_SECONDS_KEY_PREFIX = "-timeout=";
    public static final String MAX_ALLOWED_FILES_IN_JAR_KEY_PREFIX = "-max_files=";
    public static final String DECOMPILER_OUTPUT_FILE_NAME_KEY_PREFIX = "-target_file=";
    private static final int DEFAULT_DECOMPILER_TIMEOUT_SECONDS = 60;
    private static final int DEFAULT_MAX_ALLOWED_FILES_IN_JAR = 4500;
    private static final String DEFAULT_OUTPUT_FILE_NAME = "decompiled.jar";

    /**
     * The main method to run the decompiler.
     *
     * @param args Command line arguments where the second-to-last argument is the path to the input
     *     JAR file and the last argument is the path to the output directory.
     */
    public static void main(String[] args) throws IOException {
        if (args.length < 2) {
            System.err.println("Should pass input jar and target directory.");
            System.exit(1);
        }

        // Parse command-line arguments
        int timeout = DEFAULT_DECOMPILER_TIMEOUT_SECONDS;
        int maxFiles = DEFAULT_MAX_ALLOWED_FILES_IN_JAR;
        String outputJarPath = DEFAULT_OUTPUT_FILE_NAME;
        List<String> remainingArgs = new ArrayList<>();

        for (String arg : args) {
            if (arg.startsWith(DECOMPILER_TIMEOUT_SECONDS_KEY_PREFIX)) {
                timeout =
                        parseIntArgument(
                                arg, DECOMPILER_TIMEOUT_SECONDS_KEY_PREFIX, DEFAULT_DECOMPILER_TIMEOUT_SECONDS);
            } else if (arg.startsWith(MAX_ALLOWED_FILES_IN_JAR_KEY_PREFIX)) {
                maxFiles =
                        parseIntArgument(
                                arg, MAX_ALLOWED_FILES_IN_JAR_KEY_PREFIX, DEFAULT_MAX_ALLOWED_FILES_IN_JAR);
            } else if (arg.startsWith(DECOMPILER_OUTPUT_FILE_NAME_KEY_PREFIX)) {
                outputJarPath = arg.substring(DECOMPILER_OUTPUT_FILE_NAME_KEY_PREFIX.length());
            } else {
                remainingArgs.add(arg);
            }
        }

        if (outputJarPath.isEmpty() || remainingArgs.isEmpty()) {
            System.err.println("Should pass input jar and output options.");
            System.exit(1);
        }

        args = remainingArgs.toArray(new String[0]);

        String jarFilePath = args[args.length - 1];
        // create temp dir
        Path tmpDir = Files.createTempDirectory(UUID.randomUUID().toString());

        // Validate the JAR file
        try {
            long fileCount = countFilesInJar(jarFilePath);
            if (fileCount > maxFiles) {
                System.err.println(
                        "JAR "
                                + jarFilePath
                                + " contains "
                                + fileCount
                                + " files. More than allowed threshold of "
                                + maxFiles
                                + ". Creating empty JAR. Skipping decompilation.");

                // create empty jars in the output location and skip decompilation
                createEmptyJar(outputJarPath);
                return;
            }
        } catch (IOException e) {
            System.err.println("Error opening JAR file: " + e.getMessage());
            killSignal();
        }

        ScheduledExecutorService executor = Executors.newScheduledThreadPool(2);

        // Add the temporary directory to the arguments for ConsoleDecompiler, since
        // it expects the output directory as the last argument.
        String[] finalArgs =
                Stream.concat(Arrays.stream(args), Stream.of(tmpDir.toString())).toArray(String[]::new);

        Runnable task =
                () -> {
                    ConsoleDecompiler.main(finalArgs);
                };

        Runnable killSignalTask = Decompiler::killSignal;

        Future<?> future = executor.submit(task);
        executor.schedule(killSignalTask, timeout, TimeUnit.SECONDS);

        try {
            future.get(timeout, TimeUnit.SECONDS);
        } catch (TimeoutException | InterruptedException | ExecutionException e) {
            System.out.println("Task timed out.");
            // Nuke process.
            killSignal();
        } finally {
            executor.shutdownNow();
        }
        try (Stream<Path> paths = Files.walk(tmpDir)) {
            Optional<Path> decompiledJar =
                    paths
                            .filter(Files::isRegularFile)
                            .filter(path -> path.toString().endsWith(".jar"))
                            .findFirst();
            // If the decompiled jar is present, move it to the output path
            if (decompiledJar.isPresent()) {
                try {
                    Files.move(decompiledJar.get(), Paths.get(outputJarPath));
                    setCreationTimeToEpoch(outputJarPath);
                } catch (IOException e) {
                    String message = String.format("Error moving file: %s", e.getMessage());
                    System.err.println(message);
                    throw new IOException(message, e);
                }
            } else {
                // create empty jars in the output location in order to cache bazel action.
                createEmptyJar(outputJarPath);
            }
        }
    }

    /**
     * Parses an integer argument from a command-line option.
     *
     * @param arg The full command-line argument
     * @param prefix The prefix to remove
     * @param defaultValue The default value to return if parsing fails
     * @return The parsed integer value, or the default value if parsing fails
     */
    private static int parseIntArgument(String arg, String prefix, int defaultValue) {
        try {
            String valueStr = arg.substring(prefix.length());
            return Integer.parseInt(valueStr);
        } catch (NumberFormatException | IndexOutOfBoundsException e) {
            System.err.println("Invalid value for " + prefix + ", using default: " + defaultValue);
            return defaultValue;
        }
    }

    /**
     * Creates a minimal valid deterministic JAR file at the specified path. The JAR is created
     * without timestamps to ensure cache friendliness.
     *
     * @param jarFilePath The path where the empty JAR should be created
     * @throws IOException If an I/O error occurs while creating the JAR
     */
    public static void createEmptyJar(String jarFilePath) throws IOException {
        File jarFile = new File(jarFilePath);
        try (FileOutputStream fos = new FileOutputStream(jarFile);
             JarOutputStream jos = new JarOutputStream(fos)) {

            // Adding a dummy manifest to make it a valid JAR file
            ZipEntry manifestEntry = new ZipEntry("META-INF/MANIFEST.MF");
            manifestEntry.setTime(0L); // Removes timestamp
            jos.putNextEntry(manifestEntry);
            jos.write("Manifest-Version: 1.0\n".getBytes());
            jos.closeEntry();
        }
        setCreationTimeToEpoch(jarFilePath);
    }

    private static void setCreationTimeToEpoch(String filepath) throws IOException {
        Path path = Paths.get(filepath);
        FileTime epochTime = FileTime.fromMillis(0L);
        Files.getFileAttributeView(path, BasicFileAttributeView.class)
                .setTimes(epochTime, epochTime, epochTime);
    }

    /**
     * Counts the number of files (non-directory entries) in a JAR file.
     *
     * @param jarFilePath The path to the JAR file
     * @return The number of files in the JAR
     * @throws IOException If an I/O error occurs while reading the JAR file
     */
    public static long countFilesInJar(String jarFilePath) throws IOException {
        try (JarFile jarFile = new JarFile(jarFilePath)) {
            return jarFile.stream()
                    .filter(entry -> !entry.isDirectory() && entry.getName().endsWith(".class"))
                    .count();
        }
    }

    /** Kills the process. */
    private static void killSignal() {
        // Exit code 0 will make sure that bazel action is cached. We don't want to run this again if
        // it's already failed. This also should pull timed out action from the remote cache.
        System.exit(0);
    }
}
