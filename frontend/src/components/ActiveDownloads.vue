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
              >{{ formatBytes(item.bytes_completed) }} / {{ formatBytes(item.size) }} @ {{ formatBytes(item.bytes_per_second) }}/s</small>
            </td>
          </tr>
          <tr>
            <td colspan="2" class="pt-0">
              <div class="progress" style="height: 2px;">
                <div
                  class="progress-bar progress-bar-striped progress-bar-animated"
                  role="progressbar"
                  v-bind:style="{ width: item.progress + '%' }"
                  v-bind:class="{ 'bg-info': item.progress < 100, 'bg-success': item.progress == 100 }"
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
import { Complete } from "../../wailsjs/go/main/App";

export default {
  name: "ActiveDownloads",
  data() {
    return {
      downloads: [],
    };
  },
  mounted() {
    window.runtime.EventsOn("progress", (downloads) => {
        this.downloads = downloads
    })
  },
  methods: {
    formatBytes: function (value) {
      console.log(value, "format bytes")
      if (!value) {
        return;
      }
      return prettyBytes(value);
    },
    completeDownload: function (value) {
      Complete(parseInt(value)).then(() => {});
    },
  },
};
</script>
