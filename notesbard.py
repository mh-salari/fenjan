# https://notesbard.com/find-fully-phd-programs/
# https://stackoverflow.com/questions/4000508/regex-that-will-capture-everything-between-two-characters-including-multiline-bl
from bs4 import BeautifulSoup
import requests
import re
from tqdm import tqdm
from dataclasses import dataclass


@dataclass
class Number:
    title: str
    summery: str


def find_positions_url():
    print("Finding url of all phd programs advertised in the first page")
    base_url = "https://notesbard.com/find-fully-phd-programs/"
    r = requests.get(base_url)
    if r.status_code != 200:
        raise Exception(f"Request Error: {r.status_code}")

    urls = [
        a["href"]
        for a in BeautifulSoup(r.text, "html.parser").find_all(
            "a", {"class", "elementor-post__read-more"}
        )
    ]
    return urls


def get_positions(url):
    r = requests.get(url)
    if r.status_code != 200:
        raise Exception(f"Request Error: {r.status_code}")

    soup = BeautifulSoup(r.text, "html.parser")
    content = soup.find(id="content")
    regex_pattern = r'<h1 style="text-align: justify;">([\s\S]*?)<p> </p>'

    matches = re.findall(regex_pattern, str(content))

    for position in matches:

        soup = BeautifulSoup(
            '<h1 style="text-align: justify;">' + position, "html.parser"
        )
        print(soup.find("h1").text)
        print(soup.find("h2").text)

        summary = [p for p in soup.find_all("p")]

        [print(t.text) for t in summary[:-1]]
        print("-" * 50)


if __name__ == "__main__":

    urls = find_positions_url()
    for url in tqdm(urls[:2]):
        get_positions(url)
