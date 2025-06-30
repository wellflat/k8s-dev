from openai import OpenAI

# OpenAIクライアントを初期化
# vLLMサーバーのURLをbase_urlに指定します。
# vLLMは認証を必要としないため、api_keyはダミーの文字列で問題ありません。
BASE_URL="http://192.168.49.2:30201/v1"
client = OpenAI(base_url=BASE_URL, api_key="vllm")

# vLLMサーバー起動時に指定したモデル名
# run_docker.shで --model elyza/Llama-3-ELYZA-JP-8B と設定されていることを想定
#MODEL_NAME = "elyza/Llama-3-ELYZA-JP-8B"
MODEL_NAME = "meta-llama/Llama-3.1-8B-Instruct"
# Llama-3.1-8Bだと以下のエラーが出る
#'As of transformers v4.44, default chat template is no longer allowed, so you must provide a chat template if the tokenizer does not define one. None

def request_chat_completion_non_streaming():
    """
    Chat Completions APIを非ストリーミングモードで呼び出すサンプル。
    レスポンス全体を一度に受け取ります。
    """
    print("--- Chat Completions 非ストリーミングリクエスト ---")
    
    try:
        content = "hello"
        chat_completion = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "user", "content": content}
            ],
            max_tokens=50,
            temperature=0.7,
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

def request_chat_completion_streaming():
    """
    Chat Completions APIをストリーミングモードで呼び出すサンプル。
    生成されたトークンを逐次受け取ります。
    """
    print("\n--- Chat Completions ストリーミングリクエスト ---")
    
    try:
        content = "hello"
        stream = client.chat.completions.create(
            model=MODEL_NAME,
            messages=[
                {"role": "user", "content": content }
            ],
            max_tokens=500,
            temperature=0.8,
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
    request_chat_completion_non_streaming()
    request_chat_completion_streaming()
