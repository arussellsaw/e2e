<template>
  <div id="app">
			<nav class="navbar" role="navigation" aria-label="main navigation">
				<div class="navbar-brand">
					<a class="navbar-item" href="https://avocet.io">
						<p class="title" style="padding:0.5rem; margin:0.5rem;">canary</p>
					</a>
				</div>
			</nav>
			<div class="container">
				<div class="columns is-multiline">
					<div class="column is-half" v-for="test in testData" v-bind:key="test.Name">
						<TestBox v-bind:test="test"/>
					</div>
				</div>
			</div>
  </div>
</template>

<script>
import HelloWorld from './components/HelloWorld.vue'
import TestBox from './components/TestBox.vue'

export default {
  name: 'app',
	data: function() {
		return {
			testData: [],
		}
	},
  components: {
    HelloWorld,
		TestBox
  },
	mounted: function () {
		getTestStatus(this)
		var self = this;
		setInterval(function() {getTestStatus(self)}, 10000)
	},
	methods: {
	},
}

var getTestStatus = function(self) {
			var req = new Request("/api/status")
			fetch(req).then(res => res.json())
				.then((body) => {
					self.$nextTick(function () {
						self.testData = body;
					})
				})
}
</script>

