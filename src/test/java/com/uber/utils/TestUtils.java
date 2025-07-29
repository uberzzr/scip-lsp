package com.uber.utils;

import java.nio.file.Path;

public class TestUtils {
    public static Path getResourcePath(String resource) {
        String runfilesDir = System.getenv("TEST_SRCDIR");
        String workspaceName = System.getenv("TEST_WORKSPACE");
        return Path.of(runfilesDir, workspaceName, resource);
    }
}
