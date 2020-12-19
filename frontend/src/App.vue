<template>
  <div id="app" class="container">
    <div class="p-4 p-md-5 text-white bg-dark" v-if="showConfigWindow">
      <div class="col-md-6 px-0 text-white">
        <div class="form-group">
          <label for="endpoint">GRPC Server Endpoint</label>
          <input
            type="email"
            class="form-control"
            id="endpoint"
            aria-describedby="emailHelp"
            v-model="endpoint"
          />
        </div>
        <button type="button" class="btn btn-primary" @click="saveConfig">Save</button>&nbsp;
        <button type="button" class="btn btn-danger" @click="showConfig(false)">Cancel</button>
      </div>
    </div>

    <header class="blog-header py-3">
      <div class="row flex-nowrap justify-content-between align-items-center">
        <div class="col-4">
          <a class="blog-header-logo text-dark" href="#">Downloads</a>
        </div>
        <div class="col-4 d-flex justify-content-end align-items-center">
          <a class="text-muted" aria-label="Setup" @click="showConfig(true)">
            <span class="oi oi-cog" title="Setup" aria-hidden="true"></span>
          </a>
        </div>
      </div>
    </header>

    <ActiveDownloads />

    <div class="row">
      <Search />
    </div>
    <div class="row">
      <Log />
    </div>
  </div>
</template>


<script>
import "./assets/css/bootstrap.css";
import "./assets/css/open-iconic-bootstrap.css";
import "./assets/css/main.css";

import ActiveDownloads from "./components/ActiveDownloads";
import Search from "./components/Search";
import Log from "./components/Log";

export default {
  name: "App",
  components: {
    ActiveDownloads,
    Search,
    Log,
  },
  data() {
    return {
      endpoint: "",
      showConfigWindow: true,
    };
  },
  created() {
    window.backend.Agent.GetEndpoint().then((endpoint) => {
      this.endpoint = endpoint;
      if (this.endpoint != "") {
        this.showConfig(false);
      }
    });
  },
  methods: {
    showConfig: function (show) {
      this.showConfigWindow = show;
    },
    saveConfig: function () {
      window.backend.Agent.SaveEndpoint(this.endpoint).then(() => {
        this.showConfig(false);
      });
    },
  },
};
</script>
