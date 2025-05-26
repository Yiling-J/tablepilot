# Story Generation Workflow

This workflow automates the creation of characters and a story outline based on a user-selected country.

## Workflow Variables

*   `country`: (String) Defines the country of the story.

## Workflow Steps

1.  **Select a country**
    *   The user can set the `story_theme` and `num_characters` variables to customize the generation process.

2.  **Create Characters Table:**
    *   A table is created to store character information. The table name is dynamically generated based on the country.
    *   The schema for this table is defined in `characters.json` and includes:
        *   `Name`: (String) The character's name.
        *   `Archetype`: (String) The character's archetype (e.g., "Hero", "Mentor", "Villain").
        *   `Backstory`: (String) An AI-generated backstory based on the character's name, archetype, and the story theme.

3.  **Generate Characters:**
    *   The workflow generates 3 characters based on the defined theme and stores them in the characters table.

4.  **Create Story Outline Table:**
    *   A table is created to store the story outline in the country for each character.
    *   The schema for this table is defined in `story_outline.json` and includes:
        *   `Chapter`: Name of the chapter.
		*   `Character`: character of the chapter. Pick from characters table.
        *   `Location`: (String) An AI-generated location suitable for the chapter and theme.
        *   `PlotSummary`: (String) An AI-generated summary for the chapter, incorporating generated characters and aligning with the theme.

5.  **Generate Story Outline:**
    *   The workflow generates 3 chapters for the story outline, on chapter for each character.

6.  **Export Data:**
    *   Both the characters table and the story outline table are exported as CSV files.

## Files

*   `workflow.json`: Defines the main workflow steps, variables, and configurations.
*   `characters.json`: Specifies the schema for the characters table.
*   `story_outline.json`: Specifies the schema for the story outline table.
*   `README.md`: This file, providing an overview of the workflow.
