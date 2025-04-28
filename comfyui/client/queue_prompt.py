import json
import httpx
import random
import os
from typing import Dict, Any, Optional

class ComfyUIPromptGenerator:

    def __init__(self, server_address: str = "127.0.0.1:8188", workflow_path: Optional[str] = None):
        """        
        :param server_address: ComfyUIサーバーのアドレス（デフォルトは localhost の 8188 ポート）
        :param workflow_path: ワークフローJSONファイルへのパス（デフォルトはNone）
        """
        # ComfyUI サーバーのアドレスを保存
        self.server_address = server_address
        
        # workflow_path が指定されていない場合、デフォルトのパスを使用
        if workflow_path is None:
            self.workflow_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'workflow_api.json')
        else:
            self.workflow_path = workflow_path
        
        # ワークフローを読み込み、プロンプトとして保存
        self.prompt = self._load_workflow()

    def _load_workflow(self) -> Dict[str, Any]:
        """        
        :return: 読み込まれたワークフローの辞書
        :raises FileNotFoundError: ワークフローファイルが見つからない場合
        :raises ValueError: JSONの解析に失敗した場合
        """
        try:
            # ファイルを開いて JSON として読み込む
            with open(self.workflow_path, 'r') as file:
                return json.load(file)
        except FileNotFoundError:
            # ファイルが見つからない場合はエラーを発生
            raise FileNotFoundError(f"Workflow file not found: {self.workflow_path}")
        except json.JSONDecodeError:
            # JSON の解析に失敗した場合はエラーを発生
            raise ValueError(f"Invalid JSON in workflow file: {self.workflow_path}")

    def set_clip_text(self, text: str, node_id: str = "6") -> None:
        """        
        :param text: 設定するテキスト
        :param node_id: CLIPTextEncodeノードのID（デフォルト: "283"）
        """
        # 指定された node_id の inputs -> text にテキストを設定
        self.prompt[node_id]["inputs"]["text"] = text

    def set_random_seed(self, node_id: str = "294") -> None:
        """        
        :param node_id: KSamplerノードのID（デフォルト: "271"）
        """
        # 1 から 1,000,000 までのランダムな整数を生成し、シードとして設定
        self.prompt[node_id]["inputs"]["seed"] = random.randint(1, 1_000_000)

    def queue_prompt(self) -> None:
        """        
        :raises ConnectionError: サーバーへの接続に失敗した場合
        """
        data = {"prompt": self.prompt}
        with httpx.Client() as client:
            response = client.post(f"http://{self.server_address}/api/prompt", json=data, headers={'Content-Type': 'application/json'})
            if response.status_code != 200:
                raise ConnectionError(f"Failed to queue prompt: {response.text}")
            print(response.text)
        
    def generate_and_queue(self, clip_text: str) -> None:
        """        
        :param clip_text: CLIPTextEncodeノードに設定するテキスト
        """
        # CLIPTextEncodeノードのテキストを設定
        self.set_clip_text(clip_text)
        # ランダムシードを設定
        self.set_random_seed()
        # プロンプトをキューイング
        self.queue_prompt()


if __name__ == "__main__":
    custom_workflow_path = "SD3.5M_example_workflow_api.json"
    custom_generator = ComfyUIPromptGenerator(workflow_path=custom_workflow_path)
    prompt_text = "A beautiful girl in a fantasy world"
    custom_generator.generate_and_queue(prompt_text)