import type { Ref } from 'vue'

type ControllerShortcutsOptions = {
  isDisabled: Readonly<Ref<boolean>>
  loading: Readonly<Ref<boolean>>
  ready: Readonly<Ref<boolean>>
  startModal: Readonly<Ref<boolean>>
  finishModal: Readonly<Ref<boolean>>
  finishedModal: Readonly<Ref<boolean>>
  gameStarted: Readonly<Ref<boolean>>
  pickStarted: Readonly<Ref<boolean>>
  pickExhausted: Readonly<Ref<boolean>>
  qrCodeShowed: Readonly<Ref<boolean>>
  clickShowQr: () => Promise<void>
  clickHideQr: () => Promise<void>
  clickPickStart: () => Promise<void>
  clickPickFinish: () => Promise<void>
  clickPickCancel: () => Promise<void>
}

export function useControllerShortcuts(options: ControllerShortcutsOptions) {
  let lastKeyAt = 0

  function onKey(event: KeyboardEvent) {
    if (event.isComposing || event.ctrlKey || event.metaKey || event.altKey) return
    const key = event.key.toLowerCase()
    // Let focused controls keep native Enter/Space activation and text editing behavior.
    if (
      [' ', 'enter'].includes(key) &&
      event.target instanceof Element &&
      event.target.closest(
        'button, a, [role="button"], summary, input, textarea, select, [contenteditable]:not([contenteditable="false"])',
      )
    )
      return
    const toggle = [' ', 'enter', 'pagedown', 'arrowright', 'pageup', 'arrowleft'].includes(key)
    const cancel = ['escape', 'backspace'].includes(key)
    if (!toggle && !cancel && key !== 'q') return
    if (
      options.startModal.value ||
      options.finishModal.value ||
      options.finishedModal.value ||
      document.querySelector('[role="dialog"]')
    )
      return
    event.preventDefault()
    if (
      event.repeat ||
      Date.now() - lastKeyAt < 400 ||
      options.isDisabled.value ||
      options.loading.value ||
      !options.ready.value
    )
      return
    lastKeyAt = Date.now()
    if (key === 'q')
      void (options.qrCodeShowed.value ? options.clickHideQr() : options.clickShowQr())
    else if (options.gameStarted.value && !options.pickExhausted.value) {
      if (cancel && options.pickStarted.value) void options.clickPickCancel()
      else if (toggle)
        void (options.pickStarted.value ? options.clickPickFinish() : options.clickPickStart())
    }
  }

  function attach() {
    window.addEventListener('keydown', onKey)
  }

  function detach() {
    window.removeEventListener('keydown', onKey)
  }

  return { attach, detach }
}
