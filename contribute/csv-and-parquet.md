Tablepilot supports **CSV** and **Parquet** formats as data sources, as both are tabular formats. You can use local CSV and Parquet files, but you can access two widely used remote datasets:

- **Kaggle** for CSV files
- **Hugging Face Datasets** for Parquet files

When generating rows, you can randomly pick a row from these data sources as context. For example, if you're generating a recipe table, you can first assign a country to each row using a country dataset.

Below are details on how Tablepilot handles different dataset sources.

## Local File Paths

Both local **CSV** and **Parquet** files are specified using the `paths` parameter, which is an array of strings. Each path should be relative to `{source_data_dir}`, such as:

- `"test.csv"`
- `"data/test/foo.csv"`

You can use wildcards (`*` and `**`) in paths, for example:

- `"data/*.csv"` (matches all CSV files in the `data` folder)
- `"**/train*.parquet"` (matches all Parquet files starting with "train" in any subdirectory)

Tablepilot uses [`doublestar`](https://github.com/bmatcuk/doublestar) for pattern matching. You can refer to its documentation for more details.

### How Tablepilot Finds Files

1. Tablepilot starts at `{root_data_dir}`.
2. It loops through each path in the `paths` array.
3. It calls `doublestar`'s `Glob` method to find matching files.
4. It appends all matched files to an final files list.

The order of files matters when selecting rows. If all files are treated as one large table, rows from earlier files in the list will be on the top of the table.

## Local CSV

To efficiently pick random rows from large CSV files without excessive memory usage, Tablepilot builds an index when generating start:

- It scans all files row by row, grouping every 50 rows into an index entry.
- The index entry records the file index and the starting offset of each 50-row block.
- When selecting a row, Tablepilot:
  1. Finds the relevant index entry.
  2. Loads 50 rows into memory by offset and file index.
  3. Selects the required row.

- Currently The index is not cached or persisted, it is rebuilt each time a new generation starts. If your dataset is large and your disk is not fast, this may take some time. Future improvements may include persistent indexing to improve performance.

## Kaggle CSV

Kaggle provides a download API that delivers datasets as ZIP files. Using a Kaggle dataset in Tablepilot follows this process:

1. Download the dataset as a ZIP file to `{source_data_dir}/tablepilot_kaggle_cache/tmp`.
2. Unzip the dataset to `{source_data_dir}/tablepilot_kaggle_cache/{dataset_name}`
3. Delete the original ZIP file.
4. Read files as local CSVs.

Once downloaded, datasets are cached and can be used without an internet connection. If you download many datasets, the cache folder may grow large. A future CLI command may be added to list/delete cached datasets. Kaggle datasets can include other formats (e.g., JSON), but only CSV is supported at the moment. JSON support may be added in the future.

## Local Parquet

If you're unfamiliar with Parquet, see its [official documentation](https://github.com/apache/parquet-format?tab=readme-ov-file#file-format). Compared to CSV, Parquet files already include an internal index, so Tablepilot does not need to build one.

### How Tablepilot Handles Parquet Files

1. **Metadata Collection:** At generating start, Tablepilot reads metadata from all Parquet files (columns, row counts, etc.).
2. **Random Row Selection:** When selecting a row, Tablepilot:
   - Finds the correct row group.
   - Retrieves the exact row efficiently.

All Parquet processing is handled using [`parquet-go`](https://github.com/parquet-go/parquet-go).

If you're familiar with the Hugging Face Dataset Viewer, they use a [similar technique](https://github.com/huggingface/dataset-viewer/blob/main/libs/libcommon/src/libcommon/parquet_utils.py#L618), implemented in Python with `pyarrow`.

## Hugging Face Dataset

Hugging Face provides APIs that handle everything needs. Tablepilot simply calls the [Hugging Face rows API](https://huggingface.co/docs/dataset-viewer/en/rows) to get random row. The only requirement here is Parquet exports supported, which should be ture for most Hugging Face datasets.

### Important Considerations

1. **Internet Required:** Every row retrieval requires an API request, so an internet connection is necessary.
2. **API Rate Limits:** Hugging Face enforces rate limits.
   - Tablepilot limits API calls to 1 request every 5 seconds to comply. You can download Parquet files first and use them as local Parquet sources instead.

Hugging Face's auto-converted Parquet files are available in the `refs/convert/parquet` branch, for example:

[facebook/natural_reasoning Parquet files](https://huggingface.co/datasets/facebook/natural_reasoning/tree/refs%2Fconvert%2Fparquet).
