package com.uber.scip.extractor;

import org.jspecify.annotations.Nullable;
import org.objectweb.asm.AnnotationVisitor;
import org.objectweb.asm.ClassVisitor;
import org.objectweb.asm.MethodVisitor;
import org.objectweb.asm.Opcodes;

public class LombokClassVisitor extends ClassVisitor {
  // Lombok marks generated classes with the @Generated annotation. In compiled clsasses, this
  // annotation is represented as "Llombok/Generated;".
  public static final String LOMBOK_GENERATED = "Llombok/Generated;";
  private final LombokMethodVisitor methodVisitor = new LombokMethodVisitor();
  private boolean lombokGenerated = false;

  LombokClassVisitor() {
    super(Opcodes.ASM9);
  }

  @Override
  @Nullable
  public AnnotationVisitor visitAnnotation(String descriptor, boolean visible) {
    // Check for @Generated annotation
    if (LOMBOK_GENERATED.equals(descriptor)) {
      lombokGenerated = true;
    }
    return null;
  }

  @Override
  public MethodVisitor visitMethod(
      int access, String name, String descriptor, String signature, String[] exceptions) {
    return methodVisitor;
  }

  public boolean isLombokGenerated() {
    return this.lombokGenerated || methodVisitor.isLombokGenerated();
  }

  private static class LombokMethodVisitor extends MethodVisitor {
    private boolean lombokGenerated = false;

    LombokMethodVisitor() {
      super(Opcodes.ASM9);
    }

    @Override
    @Nullable
    public AnnotationVisitor visitAnnotation(String descriptor, boolean visible) {
      if (LOMBOK_GENERATED.equals(descriptor)) {
        lombokGenerated = true;
      }
      return null;
    }

    public boolean isLombokGenerated() {
      return lombokGenerated;
    }
  }
}
