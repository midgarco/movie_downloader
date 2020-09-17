<template>
  <div class="col mb-4 overflow-auto" style="height: 125px; max-height: 125px">
    <div class="row px-0">
      <table class="table table-sm table-borderless">
        <tbody
          v-for="(item, index) in downloads"
          :key="index"
          :title="JSON.stringify(item)"
          @dblclick="completeDownload(index)"
        >
          <tr class="pb-0">
            <td>{{ item.filename }}</td>
            <td class="text-right">
              <small
                class="font-weight-lighter"
              >{{ item.bytes_completed | formatBytes }} / {{ item.size | formatBytes }} @ {{ item.bytes_per_second | formatBytes }}/s</small>
            </td>
          </tr>
          <tr>
            <td colspan="2" class="pt-0">
              <div class="progress" style="height: 2px;">
                <div
                  class="progress-bar bg-success progress-bar-striped progress-bar-animated"
                  role="progressbar"
                  v-bind:style="{ width: item.progress + '%' }"
                  :aria-valuenow="item.progress"
                  aria-valuemin="0"
                  aria-valuemax="100"
                ></div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>
import prettyBytes from "pretty-bytes";
import * as Wails from '@wailsapp/runtime';

export default {
  name: "ActiveDownloads",
  data() {
    return {
      downloads: [],
    };
  },
  filters: {
    formatBytes: function (value) {
      if (!value) {
        return;
      }
      return prettyBytes(value);
    },
  },
  mounted() {
    Wails.Events.On("progress", (downloads) => {
        this.downloads = downloads
    })
  },
  methods: {
    completeDownload: function (value) {
      window.backend.Agent.Complete(parseInt(value)).then(() => {});
    },
  },
};
</script>
