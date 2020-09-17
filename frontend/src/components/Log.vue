<template>
  <div class="col-md-4">
    <div class="row no-gutters flex-md-row position-relative" v-if="logging">
      <div class="col d-flex flex-column position-static">
        <h3 class="mb-0">Log</h3>
        <ul class="list-unstyled mb-0 text-monospace overflow-auto log" style="max-height: 500px">
          <li v-for="(entry, index) in entries" :key="index">{{ entry }}</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "Log",
  data() {
      return {
          logging: false,
          entries: [
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
              "log entry",
          ]
      }
  }
};
</script>