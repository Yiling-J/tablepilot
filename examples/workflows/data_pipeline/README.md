# Customer Feedback Data Pipeline Workflow

This workflow processes customer feedback data from a CSV file, cleans it, and enriches it with sentiment analysis and standardized geographical information.

## Workflow Variable

*   `input_csv_file`: (File) Specifies the input CSV file containing raw customer feedback.
    *   Default: `sample_feedback.csv`

## Workflow Steps

1.  **Define Input Data:**
    *   The user can specify the `input_csv_file` variable to point to their customer feedback data.

2.  **Create Raw Feedback Table:**
    *   A table named `raw_customer_feedback` is created to hold the initial data.
    *   The schema for this table is defined in `raw_customer_feedback_schema.json`, which includes:
        *   `CustomerID`: (String) Unique identifier for the customer.
        *   `ProductName`: (String) Name of the product being reviewed.
        *   `FeedbackText`: (String) The raw text of the customer's feedback.
        *   `Region`: (String) The geographical region of the customer (can be inconsistent).

3.  **Import Feedback Data:**
    *   Data from the specified `input_csv_file` is imported into the `raw_customer_feedback` table.
    *   The `truncate` option is set to `true`, meaning the table is cleared before new data is imported, ensuring a fresh dataset for each workflow run.

4.  **Clean Region Data:**
    *   A new column named `CleanedRegion` (String) is added to the table.
    *   This column is then autofilled by standardizing the values from the original `Region` column. For example, variations like "US", "U.S.A", "United States", and "United States of America" are all converted to "USA". Similarly, "Ca" and "CALIFORNIA" become "CA". This ensures consistency in regional data.

5.  **Analyze Sentiment:**
    *   A new column named `Sentiment` (String) is added with predefined options: "Positive", "Neutral", "Negative".
    *   This column is autofilled by an AI model that analyzes the `FeedbackText` to classify its sentiment.

6.  **Export Cleaned Data:**
    *   The processed `raw_customer_feedback` table, now containing the cleaned region and sentiment analysis, is exported to a new CSV file named `cleaned_customer_feedback.csv`.

## Files

*   `workflow.json`: Defines the main workflow steps, variables, and configurations.
*   `raw_customer_feedback_schema.json`: Specifies the schema for the initial `raw_customer_feedback` table.
*   `sample_feedback.csv`: An example input CSV file demonstrating the expected data format and including some messy region data for cleaning.
*   `README.md`: This file, providing an overview of the data pipeline workflow.

## How to Use

1.  Ensure your customer feedback data is in a CSV file with columns matching (or adaptable to) "CustomerID", "ProductName", "FeedbackText", and "Region".
2.  If your input file is not named `sample_feedback.csv`, update the `input_csv_file` variable in `workflow.json` to point to your file.
3.  Run the workflow.
4.  The cleaned and enriched data will be available in `cleaned_customer_feedback.csv` in the workflow's output directory.
