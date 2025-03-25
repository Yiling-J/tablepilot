This example selects 30 random questions from a Parquet dataset and generates answers.

### Download Parquet Files from the Hugging Face Dataset

1. Visit [this Hugging Face dataset page](https://huggingface.co/datasets/Congliu/Chinese-DeepSeek-R1-Distill-data-110k/tree/refs%2Fconvert%2Fparquet/default/train).
2. Download one or both of the available Parquet files.
3. Create a folder named `data` in the current directory(this example dir).
4. Move the downloaded Parquet file(s) into the `data` folder.

### Create a Table

Run the following command to create a table:

```shell
tablepilot create examples/chinese_qa_parquet/table.json
```

### Generate Data for the Table

To generate question-and-answer pairs, execute:

```shell
tablepilot generate question_and_answer -c=30 -b=2
```
