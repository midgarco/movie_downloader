<template>
  <div id="app" class="container">
    <header class="blog-header py-3">
      <div class="row flex-nowrap justify-content-between align-items-center">
        <div class="col-4">
          <a class="blog-header-logo text-dark" href="#">Downloads</a>
        </div>
        <div class="col-4 d-flex justify-content-end align-items-center">
          <a class="text-muted" href="#" aria-label="Setup">
            <span class="oi oi-cog" title="Setup" aria-hidden="true"></span>
          </a>
        </div>
      </div>
    </header>

    <ActiveDownloads/>

    <div class="row">
      <Search/>

      <Log/>
    </div>
  </div>
</template>


<script>
import "./assets/css/bootstrap.css";
import "./assets/css/open-iconic-bootstrap.css";
import "./assets/css/main.css";

import ActiveDownloads from "./components/ActiveDownloads"
import Search from "./components/Search"
import Log from "./components/Log"

export default {
  name: "App",
  components: {
    ActiveDownloads,
    Search,
    Log
  }
};
</script>
