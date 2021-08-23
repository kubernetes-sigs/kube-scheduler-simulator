<template>
  <v-expansion-panels accordion multiple>
    <v-expansion-panel v-if="filterTableData.length > 1">
      <v-expansion-panel-header> Filter </v-expansion-panel-header>
      <v-expansion-panel-content>
        <v-data-table
          dense
          :headers="filterTableHeader"
          :items="filterTableData"
          item-key="Node"
        >
        </v-data-table>
      </v-expansion-panel-content>
    </v-expansion-panel>
    <v-expansion-panel v-if="scoreTableData.length > 1">
      <v-expansion-panel-header> Score </v-expansion-panel-header>
      <v-expansion-panel-content>
        <v-data-table
          dense
          :headers="scoreTableHeader"
          :items="scoreTableData"
          item-key="Node"
        >
        </v-data-table>
      </v-expansion-panel-content>
    </v-expansion-panel>
    <v-expansion-panel v-if="finalscoreTableData.length > 1">
      <v-expansion-panel-header>
        Final Score (Normalized + Applied plugin weight)
      </v-expansion-panel-header>
      <v-expansion-panel-content>
        <v-data-table
          dense
          :headers="finalscoreTableHeader"
          :items="finalscoreTableData"
          item-key="Node"
        >
        </v-data-table>
      </v-expansion-panel-content>
    </v-expansion-panel>
  </v-expansion-panels>
</template>
<script lang="ts">
import {
  ref,
  defineComponent,
  inject,
  computed,
  watch,
} from '@nuxtjs/composition-api'
import { extractTableHeader, schedulingResultToTableData } from '../lib/util'
import PodStoreKey from '../StoreKey/PodStoreKey'

export default defineComponent({
  setup() {
    const podstore = inject(PodStoreKey)
    if (!podstore) {
      throw new Error(`${PodStoreKey} is not provided`)
    }

    // scheduling results
    const filterTableHeader = ref(
      [] as Array<{
        text: string
        value: string
      }>
    )
    const filterTableData = ref(
      [] as Array<{ [name: string]: string | number }>
    )
    const scoreTableHeader = ref(
      [] as Array<{
        text: string
        value: string
      }>
    )
    const scoreTableData = ref([] as Array<{ [name: string]: string | number }>)
    const finalscoreTableHeader = ref(
      [] as Array<{
        text: string
        value: string
      }>
    )
    const finalscoreTableData = ref(
      [] as Array<{ [name: string]: string | number }>
    )

    const filterResultAnnotationKey = 'scheduler-simulator/filter-result'
    const scoreResultAnnotationKey = 'scheduler-simulator/score-result'
    const finalScoreResultAnnotationKey =
      'scheduler-simulator/finalscore-result'

    const pod = computed(() => podstore.selected)
    watch(pod, () => {
      if (pod.value?.item.metadata?.annotations) {
        if (scoreResultAnnotationKey in pod.value.item.metadata.annotations) {
          var score = JSON.parse(
            pod.value?.item.metadata?.annotations[scoreResultAnnotationKey]
          )
        }
        if (
          finalScoreResultAnnotationKey in pod.value.item.metadata.annotations
        ) {
          var finalscore = JSON.parse(
            pod.value?.item.metadata?.annotations[finalScoreResultAnnotationKey]
          )
        }
        if (filterResultAnnotationKey in pod.value.item.metadata.annotations) {
          var filter = JSON.parse(
            pod.value?.item.metadata?.annotations[filterResultAnnotationKey]
          )
        }

        filterTableHeader.value = extractTableHeader(filter)
        filterTableData.value = schedulingResultToTableData(filter)
        scoreTableHeader.value = extractTableHeader(score)
        scoreTableData.value = schedulingResultToTableData(score)
        finalscoreTableHeader.value = extractTableHeader(finalscore)
        finalscoreTableData.value = schedulingResultToTableData(finalscore)
      }
    })

    return {
      filterTableHeader,
      filterTableData,
      scoreTableHeader,
      scoreTableData,
      finalscoreTableHeader,
      finalscoreTableData,
    }
  },
})
</script>
