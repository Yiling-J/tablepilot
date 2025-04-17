#### Sources

A sourc object from which `pick`-type columns can select values. Tablepilot currently 6 types of sources:

- **ai**: Uses AI to generate a list of options dynamically. Each time a new generation starts (via the `generate` command, generate API call, or start button in UI), the options will be regenerated.
- **list**: Uses a predefined list of options.
- **linked**: Uses rows from another table as the source.
- **csv**: Uses rows from one or more CSV files / Kaggle datasets as the source. All CSV files must have a header row with column names and share the same column structure.
- **parquet**:  Uses rows from one or more Parquet files / Hugging Face datasets as the source.
- **files**: Uses file paths as values. Each value is a relative path (to {source_data_dir}). This is especially useful for image-type columns where you want to reference image file paths directly. See icon_jokes or outfit_preview in the examples directory for reference.

Each source is an object with the following fields:

Common fields:

- **name**: The name of the source (e.g., `"cuisines"`).
- **type**: The type of the source, which can be `"ai"`, `"list"`, `"linked"`, `"csv"` or `"parquet"`.

Special fields for different types:

- **ai**:
	- **prompt**: The prompt used to generate options by AI, e.g., Give me 50 common ingredients.
- **list**:
	- **options**: A list of predefined options to pick from.
	- **file**: Use file content as options, each line in the file will be one option, if this field is not empty then `options` field will be ignored. File path is **relative to the source_data_dir config**, e.g., `countries.text` or `data/countries.txt`.
- **linked**:
	- **table**: The name of the linked table.
- **csv**:
  - **paths**: A list of path patterns **relative to the source_data_dir config**. Supports exact matches, single asterisk (`*`), and double asterisk (`**`) patterns. Examples:
    - Full match: `"data/cuisines.csv"`
    - Single asterisk: `"data/*.csv"` (matches all CSV files in `data/`)
    - Double asterisk: `"data/**/*.csv"` (matches all CSV files in `data/` and subdirectories)
  - **kaggle**: The Kaggle dataset name, e.g., `"fernandol/countries-of-the-world"`.

    You can use a Kaggle dataset as a CSV data source. When doing so, Tablepilot first downloads the dataset into a cache folder:

    ```
    {source_data_dir}/tablepilot_kaggle_cache/{dataset_name (with `/` replaced by `--`)}
    ```

    Once downloaded, it functions like a local CSV source. The only difference is that the search root for `paths` is relative to the downloaded folder. The cached dataset will always be used unless you remove it manually.

    **Example:**
    ```json
    {
      "name": "countries",
      "type": "csv",
      "kaggle": "fernandol/countries-of-the-world",
      "paths": ["*.csv"]
    }
    ```

    You can find the dataset name by clicking the **Download** button in the top-right corner of the dataset page on Kaggle.

- **parquet**:
  - **paths**: Same as `CSV` source.
  - **huggingface**: If specified, uses a Hugging Face dataset and ignores `paths`.
    - **dataset**: The dataset to use (e.g., `facebook/natural_reasoning`).
    - **config**: (Optional) The dataset configuration to use, defaulting to `"default"`.
    - **split**: (Optional) The dataset split to use, defaulting to `"train"`.

- **files**:
  - **paths**: Same as `CSV` source.
