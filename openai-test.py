from openai import OpenAI

client = OpenAI(
    # defaults to os.environ.get("OPENAI_API_KEY")
    api_key="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",
)

chat_completion = client.chat.completions.create(
    messages=[
        {
            "role": "user",
            "content": "日本の首都は何？単語だけを答えて",
        }
    ],
    model="gpt-3.5-turbo",
)
print(chat_completion)