# Antibot Ticket Service

This service contains the 2 methods needed to generate cookies successfully

- `Deobfuscate` this method should be called 5-15seconds prior to the drop, it returns a hash which will be used later in the task
with **the same proxy**
  
- `Generate` this method should be called only when exactly needed, ticket cookies last around 750ms, this should only include the 
hash retrieved by calling deobfuscate