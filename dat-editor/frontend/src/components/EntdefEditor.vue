<script setup>
import { computed, useId } from 'vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'

const model = defineModel({ type: Object, default: null })

defineEmits(['save', 'delete'])

const titleId = useId()
const initialStateId = useId()

const stateKeys = computed(() => Object.keys(model.value?.states || {}))

const solidOptions = [
  { value: false, label: 'No' },
  { value: true, label: 'Yes' },
]

function addState () {
  const name = prompt('State key (e.g. idle):')
  if (!name || !model.value) {
    return
  }
  model.value.states[name] = { name }
}

function removeState (key) {
  if (!model.value) {
    return
  }
  delete model.value.states[key]
}
</script>

<template>
  <div v-if="!model" class="empty">
    Select an entity definition or create a new one.
  </div>
  <FormLayout v-else @submit.prevent="$emit('save')">
    <FormField
      label="Title"
      :for="titleId"
      description="Must match tileset type/class."
    >
      <input :id="titleId" v-model="model.title" type="text">
    </FormField>

    <FormField label="Solid" component-has-label>
      <RadioGroup
        v-model="model.solid"
        variant="boolean"
        :options="solidOptions"
      />
    </FormField>

    <FormField label="Initial state" :for="initialStateId">
      <select :id="initialStateId" v-model="model.initialState">
        <option v-for="key in stateKeys" :key="key" :value="key">
          {{ key }}
        </option>
      </select>
    </FormField>

    <div class="form-span-2">
      <FormField
        label="States"
        description="Visuals come from tileset tile type/class matching the entdef title (and title_state for non-initial states)."
      >
        <table class="states-table">
          <thead>
            <tr>
              <th>Key</th>
              <th>Display name</th>
              <th>Tileset match</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="key in stateKeys" :key="key">
              <td>{{ key }}</td>
              <td>
                <input v-model="model.states[key].name" type="text">
              </td>
              <td class="hint">
                {{ model.title || '…' }}_{{ key }} or {{ model.title || '…' }}
              </td>
              <td>
                <button type="button" class="danger" @click="removeState(key)">
                  Remove
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <button type="button" @click="addState">
          Add state
        </button>
      </FormField>
    </div>

    <template #actions>
      <button type="submit">
        Save
      </button>
      <button type="button" class="danger" @click="$emit('delete')">
        Delete
      </button>
    </template>
  </FormLayout>
</template>
