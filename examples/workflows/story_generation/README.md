# Story Generation Workflow

This workflow automates the creation of characters and a story outline based on a user-defined theme.

## Workflow Variables

*   `story_theme`: (String) Defines the genre or theme of the story (e.g., "mystery", "adventure", "sci-fi").
    *   Default: "mystery"
*   `num_characters`: (Integer) Specifies the number of characters to be generated for the story.
    *   Default: 2

## Workflow Steps

1.  **Define Story Theme and Character Count:**
    *   The user can set the `story_theme` and `num_characters` variables to customize the generation process.

2.  **Create Characters Table:**
    *   A table is created to store character information. The table name is dynamically generated based on the `story_theme` (e.g., `mystery_characters`).
    *   The schema for this table is defined in `characters.json` and includes:
        *   `Name`: (String) The character's name.
        *   `Archetype`: (String) The character's archetype (e.g., "Hero", "Mentor", "Villain").
        *   `Backstory`: (String) An AI-generated backstory based on the character's name, archetype, and the story theme.

3.  **Generate Characters:**
    *   The workflow generates `{{.num_characters}}` characters based on the defined theme and stores them in the characters table.

4.  **Create Story Outline Table:**
    *   A table is created to store the story outline. The table name is dynamically generated based on the `story_theme` (e.g., `mystery_outline`).
    *   The schema for this table is defined in `story_outline.json` and includes:
        *   `Chapter`: (Integer) The chapter number (auto-incremented).
        *   `Location`: (String) An AI-generated location suitable for the chapter and theme.
        *   `PlotSummary`: (String) An AI-generated summary for the chapter, incorporating generated characters and aligning with the theme.

5.  **Generate Story Outline:**
    *   The workflow generates 5 chapters for the story outline. Each chapter's plot summary is designed to use information from the previously generated characters table, ensuring consistency.

6.  **Export Data:**
    *   Both the characters table and the story outline table are exported as CSV files.
    *   File names are dynamic, based on the `story_theme` (e.g., `mystery_characters.csv`, `mystery_outline.csv`).

## Files

*   `workflow.json`: Defines the main workflow steps, variables, and configurations.
*   `characters.json`: Specifies the schema for the characters table.
*   `story_outline.json`: Specifies the schema for the story outline table.
*   `README.md`: This file, providing an overview of the workflow.

## How to Use

1.  Modify the `story_theme` and `num_characters` variables in `workflow.json` to suit your desired story.
2.  Run the workflow.
3.  The generated characters and story outline will be available as CSV files in the workflow's output directory.
