<template>
  <v-expansion-panels accordion multiple>
    <v-expansion-panel v-if="filterTableData().length > 1">
      <v-expansion-panel-header> Filter </v-expansion-panel-header>
      <v-expansion-panel-content>
        <v-data-table
          dense
          :headers="filterTableHeader()"
          :items="filterTableData()"
          item-key="Node"
        >
        </v-data-table>
      </v-expansion-panel-content>
    </v-expansion-panel>
    <v-expansion-panel v-if="scoreTableData().length > 1">
      <v-expansion-panel-header> Score </v-expansion-panel-header>
      <v-expansion-panel-content>
        <v-data-table
          dense
          :headers="scoreTableHeader()"
          :items="scoreTableData()"
          item-key="Node"
        >
        </v-data-table>
      </v-expansion-panel-content>
    </v-expansion-panel>
    <v-expansion-panel v-if="finalscoreTableData().length > 1">
      <v-expansion-panel-header>
        Final Score (Normalized + Applied plugin weight)
      </v-expansion-panel-header>
      <v-expansion-panel-content>
        <v-data-table
          dense
          :headers="finalscoreTableHeader()"
          :items="finalscoreTableData()"
          item-key="Node"
        >
        </v-data-table>
      </v-expansion-panel-content>
    </v-expansion-panel>
  </v-expansion-panels>
</template>
<script lang="ts">
import { V1Pod } from "@kubernetes/client-node";
import { defineComponent } from "@nuxtjs/composition-api";
import { extractTableHeader, schedulingResultToTableData } from "../lib/util";

export default defineComponent({
  props: {
    selected: {
      type: Object,
    },
  },
  
  setup(props) {
    const filterResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/filter-result";
    const scoreResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/score-result";
    const finalScoreResultAnnotationKey =
      "kube-scheduler-simulator.sigs.k8s.io/finalscore-result";

    // scheduling results
    const filterTableHeader = ():Array<{text: string;value: string;}> => {
      const p = props.selected as V1Pod;
      if (p.metadata?.annotations) {
        if (filterResultAnnotationKey in p.metadata.annotations) {
          return extractTableHeader(JSON.parse(
            p.metadata.annotations[filterResultAnnotationKey]
          ));
        }
      }
      return [];
    }

    const filterTableData = ():Array<{ [name: string]: string | number }> => {
      const p = props.selected as V1Pod;
      if (p.metadata?.annotations) {
        if (filterResultAnnotationKey in p.metadata.annotations) {
          return schedulingResultToTableData(JSON.parse(p.metadata.annotations[filterResultAnnotationKey]));
        }
      }
      return [];
    };

    const scoreTableHeader = ():Array<{text: string;value: string;}> => {
      const p = props.selected as V1Pod;
      if (p.metadata?.annotations) {
        if (scoreResultAnnotationKey in p.metadata.annotations) {
          return extractTableHeader(JSON.parse(p.metadata.annotations[scoreResultAnnotationKey]));
        }
      }
      return [];
    };

    const scoreTableData = ():Array<{ [name: string]: string | number }> => {
      const p = props.selected as V1Pod;
      if (p.metadata?.annotations) {
        if (scoreResultAnnotationKey in p.metadata.annotations) {
          return schedulingResultToTableData(JSON.parse(p.metadata.annotations[scoreResultAnnotationKey]));
        }
      }
      return [];
    };

    const finalscoreTableHeader = ():Array<{text: string;value: string;}> => {
      const p = props.selected as V1Pod;
      if (p.metadata?.annotations) {
        if (finalScoreResultAnnotationKey in p.metadata.annotations) {
          return extractTableHeader(JSON.parse(p.metadata.annotations[finalScoreResultAnnotationKey]));
        }
      }
      return [];
    };
    const finalscoreTableData = ():Array<{ [name: string]: string | number }> => {
      const p = props.selected as V1Pod;
      if (p.metadata?.annotations) {
        if (finalScoreResultAnnotationKey in p.metadata.annotations) {
          return schedulingResultToTableData(JSON.parse(p.metadata.annotations[finalScoreResultAnnotationKey]));
        }
      }
      return [];
    };

    return {
      filterTableHeader,
      filterTableData,
      scoreTableHeader,
      scoreTableData,
      finalscoreTableHeader,
      finalscoreTableData,
    };
  },
});
</script>
