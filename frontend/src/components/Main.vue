<script setup>
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import Toolbar from "primevue/toolbar";
import Button from "primevue/button";
const props = defineProps(["envs"]);
</script>

<template>
  <Toolbar>
    <template #start>
      <div>Enviorments</div>
    </template>

    <template #end>
      <Button
        label="import from template"
        icon="pi pi-file-import"
        severity="secondary"
        text
        outlined
        size="small"
      />
    </template>
  </Toolbar>
  <DataTable :value="props.envs">
    <Column field="key" header="key"></Column>
    <Column field="value" header="value"></Column>
  </DataTable>
</template>

<style scoped></style>
