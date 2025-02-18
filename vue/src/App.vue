<template>
  <v-app>
    <AppNavigation v-model="navigation" />

    <v-app-bar v-if="header" app absolute flat color="white" class="v-bar--underline">
      <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="pa-0 fill-height">
        <v-row align="center" class="flex-nowrap">
          <v-col cols="auto">
            <v-app-bar-nav-icon
              variant="text"
              title="Toggle navigation menu"
              @click.stop="navigation = !navigation"
            />
          </v-col>

          <v-col v-if="!searchVisible" cols="auto">
            <portal-target name="navigation"></portal-target>
          </v-col>

          <v-spacer />

          <v-col v-if="project.data" cols="auto">
            <AppSearch v-model="searchVisible" />
          </v-col>

          <v-col v-if="!user.isAuth" cols="auto" class="d-flex align-center">
            <v-btn text :to="{ name: 'Login' }" class="mr-1">Sign in</v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-app-bar>

    <v-main>
      <GlobalSnackbar />
      <GlobalConfirm />
      <router-view :date-range="dateRange" />
    </v-main>

    <v-footer v-if="footer" app absolute inset color="grey lighten-5">
      <v-container fluid>
        <v-row justify="center" align="center">
          <v-col cols="auto">
            <v-btn href="https://uptrace.dev/get/" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-help-circle-outline</v-icon>
              <span>Documentation</span>
            </v-btn>
            <v-btn href="https://uptrace.dev/opentelemetry/" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-open-source-initiative</v-icon>
              <span>OpenTelemetry</span>
            </v-btn>
            <v-btn href="https://uptrace.dev/get/instrument/" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-toy-brick-outline</v-icon>
              <span>Instrumentations</span>
            </v-btn>
            <v-btn href="https://t.me/uptrace" text rounded small>
              <v-icon small class="mr-1">mdi-message-outline</v-icon>
              <span>Telegram</span>
            </v-btn>
            <v-btn href="https://github.com/uptrace/uptrace" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-github</v-icon>
              <span>GitHub</span>
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-footer>
  </v-app>
</template>

<script lang="ts">
import { defineComponent, shallowRef, provide } from 'vue'

// Composables
import { useStorage } from '@/use/local-storage'
import { provideForceReload } from '@/use/force-reload'
import { useDateRange } from '@/use/date-range'
import { useUser } from '@/org/use-users'
import { useProject } from '@/org/use-projects'

// Components
import AppNavigation from '@/system/AppNavigation.vue'
import GlobalSnackbar from '@/components/GlobalSnackbar.vue'
import GlobalConfirm from '@/components/GlobalConfirm.vue'
import AppSearch from '@/components/AppSearch.vue'

// Misc
import { HOUR } from '@/util/fmt/date'

export default defineComponent({
  name: 'App',
  components: {
    AppNavigation,
    GlobalSnackbar,
    GlobalConfirm,
    AppSearch,
  },

  setup() {
    const navigation = useStorage('navigation-drawer', true)

    // Make these global across the app.
    provideForceReload()

    const header = shallowRef(true)
    provide('header', header)

    const footer = shallowRef(true)
    provide('footer', footer)

    const searchVisible = shallowRef(false)

    const dateRange = useDateRange()
    dateRange.changeDuration(HOUR)

    const user = useUser()
    const project = useProject()

    return {
      navigation,
      header,
      footer,
      searchVisible,
      dateRange,

      user,
      project,
    }
  },
})
</script>

<style lang="scss">
.theme--light,
.theme--dark {
  .v-bar--underline {
    border-width: 0 0 thin 0;
    border-style: solid;

    &.theme--light {
      border-bottom-color: #0000001f !important;
    }

    &.theme--dark {
      border-bottom-color: #ffffff1f !important;
    }
  }
}
</style>

<style lang="scss" scoped>
.hide-slider ::v-deep .v-slide-group__prev {
  display: none !important;
}

.hide-slider ::v-deep .v-slide-group__next {
  display: none !important;
}
</style>
