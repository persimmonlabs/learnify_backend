-- Migration 006: Seed Universal Blueprint
-- Populates the 7 Universal Modules that apply to EVERY domain.

INSERT INTO blueprint_modules 
(module_number, title_template, concept_slug, description_template, difficulty, estimated_hours, variable_schema, learning_objectives)
VALUES 

-- MODULE 1: THE ATOM
(
  1, 
  'The Atom: Encapsulating the {ENTITY}', 
  'atom_state',
  'In this module, we stop viewing the world as chaotic variables and start seeing it as structured Objects. You will define the {ENTITY} and its {STATE}.',
  'beginner',
  2,
  '{"required": ["ENTITY", "STATE"]}',
  '[
    "Identify the core data structure of your domain",
    "Refactor global variables into a single encapsulated unit",
    "Understand the difference between Identity (ID) and State (Data)"
  ]'::jsonb
),

-- MODULE 2: THE STANDARD
(
  2, 
  'The Standard: Verifying the {ENTITY}', 
  'standard_testing',
  'Before we build complex systems, we must trust our atoms. You will write the "Quality Control" logic for your {ENTITY} using Unit Tests.',
  'beginner',
  3,
  '{"required": ["ENTITY", "LOGIC"]}',
  '[
    "Write a failing test case before writing logic",
    "Define valid vs invalid states for {ENTITY}",
    "Implement guard clauses to prevent data corruption"
  ]'::jsonb
),

-- MODULE 3: THE WORKFLOW
(
  3, 
  'The Workflow: Scaling the {FLOW}', 
  'workflow_loops',
  'Doing it once is luck; doing it 1000 times is engineering. You will build a loop to process a collection of {ENTITY}s without crashing.',
  'intermediate',
  4,
  '{"required": ["ENTITY", "FLOW", "CONTAINER"]}',
  '[
    "Refactor manual repetition into a scalable loop",
    "Optimize O(n^2) nested loops into efficient O(n) pipelines",
    "Use the Accumulator Pattern to aggregate statistics"
  ]'::jsonb
),

-- MODULE 4: THE DECISION TREE
(
  4, 
  'The Decision Tree: Codifying {LOGIC}', 
  'decision_logic',
  'The real world has edge cases. You will map the precise logic that governs your domain, handling every "If" and "Else" gracefully.',
  'intermediate',
  4,
  '{"required": ["LOGIC", "EDGE_CASE"]}',
  '[
    "Map out a Truth Table for your domain rules",
    "Replace nested If/Else chains with clean State Machines",
    "Handle the {EDGE_CASE} scenario without crashing"
  ]'::jsonb
),

-- MODULE 5: THE ARCHETYPE
(
  5, 
  'The Archetype: Abstracting {INTERFACE}', 
  'archetype_interface',
  'Your system is too rigid. We will introduce Polymorphism to handle different types of {ENTITY}s using a shared Interface.',
  'advanced',
  5,
  '{"required": ["ENTITY", "INTERFACE", "VARIANT"]}',
  '[
    "Define a Contract (Interface) that multiple items satisfy",
    "Refactor your engine to accept the Interface, not the specific implementation",
    "Add a new {VARIANT} without changing the core engine code"
  ]'::jsonb
),

-- MODULE 6: THE FLOW STATE
(
  6, 
  'The Flow State: Asynchronous {FLOW}', 
  'flow_async',
  'The world does not wait for you. You will decouple your logic loop from your input loop to create a responsive, non-blocking system.',
  'advanced',
  5,
  '{"required": ["FLOW", "EVENT"]}',
  '[
    "Separate the Rendering Loop from the Physics/Logic Loop",
    "Handle user input events without blocking calculation",
    "Implement an Event Bus for system-wide communication"
  ]'::jsonb
),

-- MODULE 7: THE CAPSTONE
(
  7, 
  'The Capstone: The {DOMAIN} Simulation', 
  'capstone_integration',
  'You have the tools. Now build the system. You will create a persistent, data-driven simulation of a {DOMAIN} ecosystem.',
  'advanced',
  10,
  '{"required": ["DOMAIN", "CAPSTONE_GOAL"]}',
  '[
    "Integrate all previous modules into a single executable",
    "Persist state to a file/database",
    "Generate a report proving your system works"
  ]'::jsonb
);

-- Insert migration record
INSERT INTO schema_migrations (version, description)
VALUES ('006', 'Seed universal blueprint modules (1-7)');
```