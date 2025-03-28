package cli

const simpleRecipeRequest1 = "{\"max_completion_tokens\":6000,\"messages\":[{\"content\":[{\"text\":\"Give me 2 new rows for the table in JSON array format.\\n<TableDescription>table of recipes</TableDescription>\\nColumns of the table:\\n<Columns>\\n  <Column id=\\\"id\\\" name=\\\"id\\\" description=\\\"index of the row, always starting from 0 in each generation\\\" type=\\\"integer\\\"/>\\n  <Column id=\\\"UkLWZg\\\" name=\\\"Name\\\" description=\\\"recipe name\\\" type=\\\"string\\\"/>\\n  <Column id=\\\"gbHJdm\\\" name=\\\"Ingredients\\\" description=\\\"list of ingredients\\\" type=\\\"array\\\"/>\\n  <Column id=\\\"EfhxLZ\\\" name=\\\"Cooking Time\\\" description=\\\"time required to cook the recipe (minutes)\\\" type=\\\"integer\\\"/>\\n  <Column id=\\\"VqXmZF\\\" name=\\\"Cuisine\\\" description=\\\"type of cuisine\\\" type=\\\"string\\\"/>\\n  <Column id=\\\"uw2YK1\\\" name=\\\"Difficulty Level\\\" description=\\\"difficulty level of the recipe (1-5)\\\" type=\\\"integer\\\"/>\\n</Columns>\\nGenerate values for the following missing columns:\\n<MissingColumns>\\n  <Column id=\\\"id\\\"/>\\n  <Column id=\\\"UkLWZg\\\"/>\\n  <Column id=\\\"gbHJdm\\\"/>\\n  <Column id=\\\"EfhxLZ\\\"/>\\n  <Column id=\\\"VqXmZF\\\"/>\\n  <Column id=\\\"uw2YK1\\\"/>\\n</MissingColumns>\\n\",\"type\":\"text\"}],\"role\":\"user\"}],\"model\":\"gpt-4o\",\"presence_penalty\":0,\"response_format\":{\"json_schema\":{\"description\":\"schema for table\",\"name\":\"schema\",\"schema\":{\"properties\":{\"data\":{\"items\":{\"properties\":{\"id\":{\"type\":\"integer\"},\"UkLWZg\":{\"type\":\"string\"},\"gbHJdm\":{\"items\":{\"type\":\"string\"},\"type\":\"array\"},\"EfhxLZ\":{\"type\":\"integer\"},\"VqXmZF\":{\"type\":\"string\"},\"uw2YK1\":{\"type\":\"integer\"}},\"additionalProperties\":false,\"type\":\"object\",\"required\":[\"id\",\"UkLWZg\",\"gbHJdm\",\"EfhxLZ\",\"VqXmZF\",\"uw2YK1\"]},\"type\":\"array\"}},\"additionalProperties\":false,\"type\":\"object\"}},\"type\":\"json_schema\"},\"temperature\":0.6}"

const simpleRecipeResponse1 = `{
  "id": "chatcmpl-B9MBs8CjcvOU2jLn4n570S5qMJKcT",
  "object": "chat.completion",
  "created": 1741569952,
  "model": "gpt-4o-2024-08-06",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "{\"data\": [{\"id\":0,\"UkLWZg\":\"foo\",\"gbHJdm\":[\"ing1\",\"ing2\"],\"EfhxLZ\":10,\"VqXmZF\":\"cc\",\"uw2YK1\":1},{\"id\":1,\"UkLWZg\":\"foo1\",\"gbHJdm\":[\"ing3\",\"ing4\"],\"EfhxLZ\":12,\"VqXmZF\":\"cc1\",\"uw2YK1\":2}]}",
        "refusal": null,
        "annotations": []
      },
      "logprobs": null,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 19,
    "completion_tokens": 10,
    "total_tokens": 29,
    "prompt_tokens_details": {
      "cached_tokens": 0,
      "audio_tokens": 0
    },
    "completion_tokens_details": {
      "reasoning_tokens": 0,
      "audio_tokens": 0,
      "accepted_prediction_tokens": 0,
      "rejected_prediction_tokens": 0
    }
  },
  "service_tier": "default"
}
`

const simpleRecipeRequest2 = "{\"max_completion_tokens\":6000,\"messages\":[{\"content\":[{\"text\":\"Give me 2 new rows for the table in JSON array format.\\n<TableDescription>table of recipes</TableDescription>\\nConsider the following existing values for column UkLWZg, collected from previous rows. Try not to repeat any of these values in your output for column UkLWZg:\\n<Values column_id=\\\"UkLWZg\\\">\\n  <Value>foo1</Value>\\n  <Value>foo</Value>\\n  <Value>foo1</Value>\\n  <Value>foo</Value>\\n</Values>\\nColumns of the table:\\n<Columns>\\n  <Column id=\\\"id\\\" name=\\\"id\\\" description=\\\"index of the row, always starting from 0 in each generation\\\" type=\\\"integer\\\"/>\\n  <Column id=\\\"UkLWZg\\\" name=\\\"Name\\\" description=\\\"recipe name\\\" type=\\\"string\\\"/>\\n  <Column id=\\\"gbHJdm\\\" name=\\\"Ingredients\\\" description=\\\"list of ingredients\\\" type=\\\"array\\\"/>\\n  <Column id=\\\"EfhxLZ\\\" name=\\\"Cooking Time\\\" description=\\\"time required to cook the recipe (minutes)\\\" type=\\\"integer\\\"/>\\n  <Column id=\\\"VqXmZF\\\" name=\\\"Cuisine\\\" description=\\\"type of cuisine\\\" type=\\\"string\\\"/>\\n  <Column id=\\\"uw2YK1\\\" name=\\\"Difficulty Level\\\" description=\\\"difficulty level of the recipe (1-5)\\\" type=\\\"integer\\\"/>\\n</Columns>\\nGenerate values for the following missing columns:\\n<MissingColumns>\\n  <Column id=\\\"id\\\"/>\\n  <Column id=\\\"UkLWZg\\\"/>\\n  <Column id=\\\"gbHJdm\\\"/>\\n  <Column id=\\\"EfhxLZ\\\"/>\\n  <Column id=\\\"VqXmZF\\\"/>\\n  <Column id=\\\"uw2YK1\\\"/>\\n</MissingColumns>\\n\",\"type\":\"text\"}],\"role\":\"user\"}],\"model\":\"gpt-4o\",\"presence_penalty\":0,\"response_format\":{\"json_schema\":{\"description\":\"schema for table\",\"name\":\"schema\",\"schema\":{\"properties\":{\"data\":{\"items\":{\"properties\":{\"id\":{\"type\":\"integer\"},\"UkLWZg\":{\"type\":\"string\"},\"gbHJdm\":{\"items\":{\"type\":\"string\"},\"type\":\"array\"},\"EfhxLZ\":{\"type\":\"integer\"},\"VqXmZF\":{\"type\":\"string\"},\"uw2YK1\":{\"type\":\"integer\"}},\"additionalProperties\":false,\"type\":\"object\",\"required\":[\"id\",\"UkLWZg\",\"gbHJdm\",\"EfhxLZ\",\"VqXmZF\",\"uw2YK1\"]},\"type\":\"array\"}},\"additionalProperties\":false,\"type\":\"object\"}},\"type\":\"json_schema\"},\"temperature\":0.6}"

const simpleRecipeResponse2 = `{
  "id": "chatcmpl-B9MBs8CjcvOU2jLn4n570S5qMJKcT",
  "object": "chat.completion",
  "created": 1741569952,
  "model": "gpt-4o-2024-08-06",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "{\"data\": [{\"id\":0,\"UkLWZg\":\"bar\",\"gbHJdm\":[\"ing-1\",\"ing-2\"],\"EfhxLZ\":3,\"VqXmZF\":\"cd\",\"uw2YK1\":2},{\"id\":1,\"UkLWZg\":\"bar1\",\"gbHJdm\":[\"ing-3\",\"ing-4\"],\"EfhxLZ\":15,\"VqXmZF\":\"cd1\",\"uw2YK1\":2}]}",
        "refusal": null,
        "annotations": []
      },
      "logprobs": null,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 19,
    "completion_tokens": 10,
    "total_tokens": 29,
    "prompt_tokens_details": {
      "cached_tokens": 0,
      "audio_tokens": 0
    },
    "completion_tokens_details": {
      "reasoning_tokens": 0,
      "audio_tokens": 0,
      "accepted_prediction_tokens": 0,
      "rejected_prediction_tokens": 0
    }
  },
  "service_tier": "default"
}
`
