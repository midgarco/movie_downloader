<template>
  <div class="col-md-8">
    <div class="row no-gutters flex-md-row mb-4 position-relative">
      <div class="col d-flex flex-column position-static">
        <h3 class="mb-0">Movie</h3>
        <div class="form-inline">
          <div class="form-group mr-sm-3 mb-2">
            <input type="text" class="form-control" v-model="query" value @keyup.enter="queryMovie" />
          </div>
          <button type="submit" class="btn btn-primary mb-2" @click="queryMovie">Search</button>
        </div>
      </div>
    </div>

    <div class="d-flex justify-content-center" v-if="loading">
      <div class="spinner-border" role="status">
        <span class="sr-only">Loading...</span>
      </div>
    </div>

    <div class="row no-gutters overflow-auto" style="max-height: 430px" v-show="results.length">
      <div class="list-group mb-2 w-100">
        <a
          class="list-group-item list-group-item-action search-result"
          v-for="result in results"
          :key="result.id"
          :title="JSON.stringify(result)"
          @dblclick="downloadMovie(result)"
        >
          <h5 class="mb-0">{{ result.filename }}</h5>
          <small class="text-muted font-weight-lighter">{{ result.post_date }}</small>

          <table class="table table-sm mt-2">
            <thead>
              <tr>
                <th scope="col">Size</th>
                <th scope="col">Resolution</th>
                <th scope="col">Runtime</th>
                <th scope="col">Codec</th>
                <th scope="col">Audios</th>
              </tr>
            </thead>
            <tbody class="font-weight-light">
              <tr>
                <td>{{ result.size }}</td>
                <td>{{ result.resolution }}</td>
                <td>{{ result.runtime }}</td>
                <td>{{ result.codec }}</td>
                <td>
                  <ul class="list-unstyled">
                    <li v-for="lang in result.audio_languages" :key="lang">{{ lang }}</li>
                  </ul>
                </td>
              </tr>
            </tbody>
          </table>
        </a>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "Search",
  data() {
    return {
      loading: false,
      query: "",
      results: [],
    };
  },
  methods: {
    queryMovie: function () {
      this.loading = true;
      this.results = [];

      var value = this.query && this.query.trim();
      if (!value) {
        this.loading = false;
        return;
      }

      window.backend.Agent.Search(value).then((resp) => {
        this.results = resp.results.movies;
        this.loading = false;
      });
    },
    downloadMovie: function (movie) {
      window.backend.Agent.Download(JSON.stringify(movie)).then(() => {
        console.log("done");
      });
    },
  },
};
</script>