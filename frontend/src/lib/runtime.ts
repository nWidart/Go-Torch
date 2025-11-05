export const hasRuntime = () => typeof (window as any).runtime !== 'undefined'

export const getBackend = () => (window as any).go?.app?.App as undefined | {
  SelectLogFile: () => Promise<string>
  StartTrackingWithOptions?: (path: string, fromStart: boolean) => Promise<void>
  StartTracking?: (path: string) => Promise<void>
  PauseSession?: () => Promise<void> | void
  ResumeSession?: () => Promise<void> | void
  Stop: () => Promise<void>
  Reset: () => Promise<void>
}
