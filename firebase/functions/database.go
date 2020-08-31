package functions

// RTDBTriggerCreate : Triggered when new data is created in the Realtime Database.
const RTDBTriggerCreate = "providers/google.firebase.database/eventTypes/ref.create"

// RTDBTriggerDelete : Triggered when data is deleted from the Realtime Database.
const RTDBTriggerDelete = "providers/google.firebase.database/eventTypes/ref.delete"

// RTDBTriggerUpdate : Triggered when data is updated in the Realtime Database.
const RTDBTriggerUpdate = "providers/google.firebase.database/eventTypes/ref.update"

// RTDBTriggerWrite : Triggered on any mutation event: when data is created, updated, or deleted in the Realtime Database.
const RTDBTriggerWrite = "providers/google.firebase.database/eventTypes/ref.write"
