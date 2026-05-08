<!--
Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
Copyright (C) 2026 Snuffy2
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
  <div id="auth">
    <div id="auth-frame">
      <div id="auth-content">
        <h1>Authentication required</h1>

        <form class="form1" action="javascript:;" method="POST" @submit="auth">
          <fieldset>
            <div
              class="field"
              :class="{
                error: passphraseErr.length > 0 || error.length > 0,
              }"
            >
              Passphrase

              <input
                v-model="passphrase"
                v-focus="true"
                :disabled="submitting"
                type="password"
                autocomplete="off"
                name="field.field.name"
                placeholder="----------"
                autofocus="autofocus"
              />

              <div
                v-if="passphraseErr.length <= 0 && error.length <= 0"
                class="message"
              >
                A valid password is required in order to use this
                <a href="https://github.com/Snuffy2/sshwifty">Sshwifty</a>
                instance
              </div>
              <div v-else class="error">
                {{ passphraseErr || error }}
              </div>
            </div>

            <div class="field">
              <button type="submit" :disabled="submitting" @click="auth">
                Authenticate
              </button>
            </div>
          </fieldset>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
/**
 * @file auth.vue
 * @description Authentication wall component. Renders a passphrase form that
 * is shown when the Sshwifty backend requires a passphrase before allowing
 * access. Emits an `"auth"` event with the passphrase on valid submission and
 * reflects server-returned errors through the `error` prop.
 */
export default {
  directives: {
    /**
     * `v-focus` directive: moves browser focus to the bound element on insert
     * when the binding value is truthy.
     */
    focus: {
      /**
       * Called after the element is inserted into the DOM.
       *
       * @param {HTMLElement} el - The element the directive is bound to.
       * @param {{ value: boolean }} binding - Directive binding; focus is applied
       *   only when `binding.value` is truthy.
       * @returns {void}
       */
      mounted(el, binding) {
        if (!binding.value) {
          return;
        }

        el.focus();
      },
    },
  },
  props: {
    /**
     * Server-returned authentication error message. When non-empty the error is
     * displayed in the form and `submitting` is reset to allow retrying.
     *
     * @type {string}
     */
    error: {
      type: String,
      default: "",
    },
  },
  emits: ["auth"],
  data() {
    return {
      submitting: false,
      passphrase: "",
      passphraseErr: "",
    };
  },
  watch: {
    error(newVal) {
      if (newVal.length > 0) {
        this.submitting = false;
      }
    },
  },
  mounted() {},
  methods: {
    /**
     * Validates the passphrase field and emits an `"auth"` event to the parent.
     *
     * Guards against empty submissions and duplicate in-flight requests via the
     * `submitting` flag. Clears `passphraseErr` before emitting.
     *
     * @returns {void}
     */
    auth() {
      if (this.passphrase.length <= 0) {
        this.passphraseErr = "Passphrase cannot be empty";

        return;
      }

      if (this.submitting) {
        return;
      }

      this.submitting = true;

      this.passphraseErr = "";

      this.$emit("auth", this.passphrase);
    },
  },
};
</script>
