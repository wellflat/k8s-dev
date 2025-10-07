from openai import OpenAI

# OpenAIクライアントを初期化
# vLLMサーバーのURLをbase_urlに指定します。
# vLLMは認証を必要としないため、api_keyはダミーの文字列で問題ありません
BASE_URL="http://localhost:30201/v1"
#BASE_URL="http://localhost:4000/v1"
client = OpenAI(base_url=BASE_URL, api_key="dummy-key")

# vLLMサーバー起動時に指定したモデル名
#MODEL_NAME = "elyza/Llama-3-ELYZA-JP-8B"
#MODEL_NAME = "meta-llama/Llama-3.1-8B-Instruct"
MODEL_NAME = "openai/gpt-oss-20b"
#MODEL_NAME = "gpt-oss-20b"

def request_chat_completion_non_streaming(prompts: str):
    """
    Chat Completions APIを非ストリーミングモードで呼び出すサンプル
    レスポンス全体を一度に受け取ります
    """
    print("--- Chat Completions 非ストリーミングリクエスト ---")
    
    try:
        chat_completion = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "user", "content": prompts}
            ],
            max_tokens=2048,
            temperature=1.0,
            stream=False
        )
        
        # 生成されたテキストを表示
        print("生成されたテキスト:")
        print(chat_completion.choices[0].message.content)
        
        # トークン使用量を表示
        print("\nトークン使用量:")
        print(chat_completion.usage)
    except Exception as e:
        print(f"リクエスト中にエラーが発生しました: {e}")

def request_chat_completion_streaming(prompts: str):
    """
    Chat Completions APIをストリーミングモードで呼び出すサンプル
    生成されたトークンを逐次受け取ります
    """
    print("\n--- Chat Completions ストリーミングリクエスト ---")
    
    try:
        stream = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "user", "content": prompts }
            ],
            max_tokens=2048,
            temperature=1.0,
            stream=True
        )

        print("生成されたテキスト:")
        # ストリームからチャンクを一つずつ処理
        for chunk in stream:
            # チャンク内のテキスト(delta)を取得して表示
            # completions APIの .text ではなく、chat completions APIの .delta.content を使用
            content = chunk.choices[0].delta.content
            if content:
                print(content, end='', flush=True)
        print() # 最後に改行

    except Exception as e:
        print(f"リクエスト中にエラーが発生しました: {e}")


if __name__ == "__main__":
    prompts = "仕事の熱意を取り戻すためのアイデアを5つ挙げてください"
    #prompts = "Find the sum of all integer bases $b>9$ for which $17_b$ is a divisor of $97_b.$"
    request_chat_completion_non_streaming(prompts)
    request_chat_completion_streaming(prompts)
