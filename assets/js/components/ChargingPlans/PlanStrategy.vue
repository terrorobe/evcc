<template>
	<div class="collapsible-wrapper" :class="{ open: show }">
		<div class="collapsible-content pb-3">
			<div v-if="optimizationDisabled" class="row mb-4">
				<div class="small text-muted">
					<strong class="text-primary">{{ $t("general.note") }}</strong>
					{{ $t("main.chargingPlan.strategyDisabledDescription") }}
				</div>
			</div>
			<div class="row">
				<div class="col-12 col-sm-6 col-lg-4 mb-3">
					<div class="row">
						<label :for="formId('continuous')" class="col-form-label col-5 col-sm-12">
							{{ $t("main.chargingPlan.optimization.label") }}
						</label>
						<div class="col-7 col-sm-12">
							<select
								:id="formId('continuous')"
								v-model="localContinuous"
								class="form-select"
								:disabled="optimizationDisabled"
								@change="updateStrategy"
							>
								<option :value="false">
									{{ $t("main.chargingPlan.optimization.cheapest") }}
								</option>
								<option :value="true">
									{{ $t("main.chargingPlan.optimization.continuous") }}
								</option>
							</select>
						</div>
					</div>
				</div>
				<div class="col-sm-6 col-lg-4 mb-3">
					<div class="row">
						<label :for="formId('precondition')" class="col-form-label col-5 col-sm-12">
							{{ $t("main.chargingPlan.precondition.label") }}
						</label>
						<div class="col-7 col-sm-12">
							<select
								:id="formId('precondition')"
								v-model="localPrecondition"
								class="form-select"
								:disabled="optimizationDisabled"
								@change="updateStrategy"
							>
								<option :value="0">
									{{ $t("main.chargingPlan.precondition.optionNo") }}
								</option>
								<option
									v-for="opt in preconditionOptions"
									:key="opt.value"
									:value="opt.value"
								>
									{{ opt.name }}
								</option>
							</select>
						</div>
					</div>
				</div>
				<div class="col-sm-6 col-lg-4 mb-3">
					<div class="row">
						<label :for="formId('power')" class="col-form-label col-5 col-sm-12">
							{{ $t("main.chargingPlan.power.label") }}
						</label>
						<div class="col-7 col-sm-12">
							<select
								:id="formId('power')"
								v-model="localPower"
								class="form-select"
								@change="updateStrategy"
							>
								<option value="max">
									{{ $t("main.chargingPlan.power.max") }}
								</option>
								<option value="required">
									{{ $t("main.chargingPlan.power.required") }}
								</option>
							</select>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import formatter from "@/mixins/formatter";
import type { PlanPowerMode, PlanStrategy } from "./types";

export default defineComponent({
	name: "ChargingPlanStrategy",
	mixins: [formatter],
	props: {
		id: [String, Number],
		show: Boolean,
		precondition: { type: Number, default: 0 },
		continuous: { type: Boolean, default: false },
		power: { type: String as () => PlanPowerMode, default: "max" },
		optimizationDisabled: Boolean,
	},
	emits: ["update"],
	data() {
		return {
			localPrecondition: this.precondition,
			localContinuous: this.continuous,
			localPower: this.power,
		};
	},
	computed: {
		preconditionOptions() {
			const HOUR = 60 * 60;
			const QUARTER_HOUR = 0.25 * HOUR;
			const HALF_HOUR = 0.5 * HOUR;
			const ONE_HOUR = 1 * HOUR;
			const TWO_HOURS = 2 * HOUR;
			const EVERYTHING = 7 * 24 * HOUR;

			const options = [QUARTER_HOUR, HALF_HOUR, ONE_HOUR, TWO_HOURS, EVERYTHING];

			// support custom values (via API)
			if (this.localPrecondition && !options.includes(this.localPrecondition)) {
				options.push(this.localPrecondition);
			}

			return options.map((s) => ({
				value: s,
				name:
					s === EVERYTHING
						? this.$t("main.chargingPlan.precondition.optionAll")
						: this.fmtDurationLong(s),
			}));
		},
	},
	watch: {
		precondition: {
			handler(newValue: number) {
				// Only update if value actually changed from external source
				if (newValue !== this.localPrecondition) {
					this.localPrecondition = newValue;
				}
			},
			immediate: true,
		},
		continuous: {
			handler(newValue: boolean) {
				// Only update if value actually changed from external source
				if (newValue !== this.localContinuous) {
					this.localContinuous = newValue;
				}
			},
			immediate: true,
		},
		power: {
			handler(newValue: PlanPowerMode) {
				if (newValue !== this.localPower) {
					this.localPower = newValue;
				}
			},
			immediate: true,
		},
	},
	methods: {
		formId(name: string) {
			return `chargingplan-${this.id}-${name}`;
		},
		updateStrategy(): void {
			const strategy: PlanStrategy = {
				continuous: this.localContinuous,
				precondition: this.localPrecondition,
				power: this.localPower,
			};
			this.$emit("update", strategy);
		},
	},
});
</script>
