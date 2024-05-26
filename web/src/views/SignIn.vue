<template>
  <div class="login-wrapper">
    <div class="login-container">
      <div class="title-container">
        <h1 class="title">Elaina</h1>
        <div class="sub-title">
          <p class="tip">Container-based remote code runner.</p>
        </div>
      </div>

      <t-form
          ref="form"
          class="item-container"
          :data="formData"
          :rules="FORM_RULES"
          label-width="0"
          @submit="onSubmit"
      >
        <t-form-item name="password">
          <t-input
              v-model="formData.password"
              size="large"
              :type="showPsw ? 'text' : 'password'"
              clearable
              placeholder="Password"
          >
            <template #prefix-icon>
              <t-icon name="lock-on"/>
            </template>
            <template #suffix-icon>
              <t-icon :name="showPsw ? 'browse' : 'browse-off'" @click="showPsw = !showPsw"/>
            </template>
          </t-input>
        </t-form-item>

        <t-form-item class="btn-container">
          <t-button block size="large" type="submit"> Sign In</t-button>
        </t-form-item>
      </t-form>
    </div>

    <footer class="copyright">
      Copyright @ 2021-{{ currentYear }} Elaina. All Rights Reserved.
      <div class="links">
        <a href="https://github.com/wuhan005/Elaina" rel="nofollow" target="_blank">
          <t-icon name="logo-github-filled"/>
        </a>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import type {FormRule, SubmitContext} from 'tdesign-vue-next';
import {MessagePlugin} from 'tdesign-vue-next';
import {ref} from 'vue';
import {useRouter} from 'vue-router';
import {type SignInForm, signIn} from "@/api/auth";
import {useAuthStore} from "@/store";

const currentYear = new Date().getFullYear();
const router = useRouter();
const authStore = useAuthStore();
const FORM_RULES: Record<string, FormRule[]> = {
  password: [{required: true, message: 'Password is required.', type: 'error'}],
};
const formData = ref<SignInForm>({password: ''});
const showPsw = ref(false);

const onSubmit = async (ctx: SubmitContext) => {
  if (ctx.validateResult === true) {

    signIn(formData.value).then(token => {
      authStore.setToken(token);

      MessagePlugin.success('Sign in successfully');
      router.push({name: 'dashboard'});
    })
  }
};
</script>

<style scoped lang="less">
.login-wrapper {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-size: cover;
  background-position: 100%;
  position: relative;
}

.login-container {
  position: absolute;
  top: 22%;
  left: 5%;
  min-height: 500px;
}

.title-container {
  .title {
    font: var(--td-font-headline-large);
    color: var(--td-text-color-primary);
    margin-top: var(--td-comp-margin-xs);

    &.margin-no {
      margin-top: 0;
    }
  }

  .sub-title {
    margin-top: var(--td-comp-margin-xl);

    .tip {
      display: inline-block;
      margin-right: var(--td-comp-margin-s);
      font: var(--td-font-body-medium);

      &:first-child {
        color: var(--td-text-color-secondary);
      }
    }
  }
}

.item-container {
  width: 400px;
  margin-top: var(--td-comp-margin-xxxxl);

  .check-container {
    display: flex;
    align-items: center;

    &.remember-pwd {
      margin-bottom: var(--td-comp-margin-l);
      justify-content: space-between;
    }

    span {
      color: var(--td-brand-color);

      &:hover {
        cursor: pointer;
      }
    }
  }

  .btn-container {
    margin-top: var(--td-comp-margin-xxxxl);
  }
}

.copyright {
  font: var(--td-font-body-medium);
  position: absolute;
  left: 5%;
  bottom: 64px;
  color: var(--td-text-color-secondary);

  .links {
    margin-top: var(--td-comp-margin-s);
    width: 100%;
    display: flex;
    justify-content: center;
  }
}

@media screen and (height <= 700px) {
  .copyright {
    display: none;
  }
}
</style>
