<template>
	<div class="has-text-white" style="border-radius: 3px; padding: 2rem;" v-bind:class="testClasses(test)">
		<h1 class="title is-4 has-text-white">
			{{ test.Name }}
		</h1>
		<div class="columns" style="height:100px;">
			<div class="column is-one-third">
				<p class="is-size-7">state: {{ test.State }}</p>
				<p class="is-size-7">pass rate: {{ Math.round(((test.Successes/(test.Successes+test.Failures)) * 100)||0) }}%</p>
			</div>
			<Log v-if="isFailingOrRunning(test)" v-bind:name="test.Name" v-bind:baseOutput="test.LastFailureOutput" v-bind:state="test.State"/>
		</div>
	</div>
</template>

<script>
import Log from './Log.vue'
export default {
  name: 'TestBox',
  props: {
		test: {}, 
  },
	components: {
		Log
	},
	methods: {
		testClasses: function (test) {
			return {
				'has-background-danger': test.State == "FAILED",
				'has-background-success': test.State == "PASSED",
				'has-background-grey-light': test.State == "RUNNING",
				'has-background-grey-dark': test.State == "",
			}
		},
		isFailing: function (test) {
			return test.State == "FAILED"
		},
		isFailingOrRunning: function (test) {
			return (test.State == "RUNNING" || test.State == "FAILED")
		}

	} 
}
</script>
