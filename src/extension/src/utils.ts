import * as os from 'os'

export function sendTelemetryEvent(command: string, data: any): void {
    // This function is a placeholder for reporting data to a telemetry provider
}

const PLATFORM_DARWIN_ARM64 = 'darwin_arm64'
const PLATFORM_DARWIN_AMD64 = 'darwin_amd64'
const PLATFORM_LINUX_AMD64 = 'linux_amd64'

/**
 * Determines the current platform for UEXEC YAML.
 * @returns The platform string.
 */
export function getPlatform(): string {
    const platform = os.platform()
    const arch = os.arch()

    if (platform === 'darwin') {
      return arch === 'arm64' ? PLATFORM_DARWIN_ARM64 : PLATFORM_DARWIN_AMD64
    } else if (platform === 'linux') {
      return PLATFORM_LINUX_AMD64
    }

    // Default to linux_amd64 if unknown
    return PLATFORM_LINUX_AMD64
  }
