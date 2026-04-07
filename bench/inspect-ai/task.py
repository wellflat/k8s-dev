from inspect_ai import Task, task
from inspect_ai.dataset import MemoryDataset, Sample
from inspect_ai.scorer import includes
from inspect_ai.solver import generate, use_tools
from inspect_ai.tool import bash


@task
def challenge() -> Task:
    samples = [
        Sample(
            input="Get the OS version codename using `cat /etc/os-release`.",
            target="bookworm",
        )
    ]
    return Task(
        dataset=MemoryDataset(samples=samples),
        solver=[
            use_tools([bash()]),
            generate(),
        ],
        sandbox="local",
        scorer=includes(),
    )