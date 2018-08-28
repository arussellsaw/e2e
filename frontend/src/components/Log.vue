<template>
	<div class="logbox has-background-grey-dark column is-two-thirds">
		<p style="white-space: pre;" class="is-size-7">{{ output }}</p>
	</div>
</template>

<script>
export default {
  name: 'Log',
	data: function() {
		return {
			output: ""
		}
	},
  props: {
		name: "",
		state: "",
		baseOutput: "",
  },
	mounted: function() {
		if (this.state != "RUNNING") {
			this.output = this.baseOutput
		}
		var self = this;
		getOutput(self);
		setInterval(function() {
			getOutput(self);
		}, 1000)
	},
	methods: {
	} 
}

var getOutput = function(self) {
	if (self.state != "RUNNING") {
		return
	}
	var req = new Request("/api/v1/namespaces/default/services/e2e/proxy/api/log/"+self.name)
	fetch(req).then(res => res.text())
		.then(function(body) {
			self.$nextTick(function() {
				self.output = body
			})
		})
}

</script>

<style scoped>
.logbox {
	overflow-y: scroll; 
	overflow-x: scroll; 
	padding: 0.5rem; 
	margin-right: 1.5rem;
	border-radius: 3px;
}
</style>
