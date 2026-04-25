class KeyboardManager {
  constructor() {
    this.keys = {
      bucket:       "w",
      pen:          "q",
      rubber:       "e",
      size8:        "1",
      size16:       "2",
      size24:       "3",
      size32:       "4",
      undo:         "z",
      
      // multiple modifiers should be separated by +, for example "ctrl+shift"
      undoModifier: "ctrl",
    }
  }

  get(key) {
    return this.keys[key]
  }
}

export default KeyboardManager;
