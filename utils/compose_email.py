#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Des 6 2022
@author: Hue (MohammadHossein) Salari
@email:hue.salari@gmail.com
"""

import re
import os
import random
from datetime import datetime
from urlextract import URLExtract
from utils.send_email import send_email


def format_position_summery_text(text):
    extractor = URLExtract()
    formatted_text = ""
    text = re.sub(r"\n+", "\n", text)
    text = text.replace("\n", " <br> ")
    words = text.split()
    for word in words:
        url = extractor.find_urls(word)
        if word[0] == "#":
            formatted_text += '<span style="color:#68677b">' + word + "</span>" + " "
        elif "&link" in word:
            link = word[word.find("&link") : word.find("*")].replace("&link", "")
            text = word[word.find("*") + 1 :].replace("%20", " ")
            formatted_text += " "
            formatted_text += f'<a href="{link}" rel="noopener" style="text-decoration: underline; color: #8a3b8f;" target="_blank">{text}</a>'
            formatted_text += " "
            formatted_text.replace(word, "")
        elif url:
            formatted_text += " "
            formatted_text += f'<a href="{url[0]}" rel="noopener" style="text-decoration: underline; color: #8a3b8f;" target="_blank">{url[0]}</a>'
            formatted_text += word.replace(url[0], "")
            formatted_text += " "
        else:
            formatted_text += word + " "

    return formatted_text


def compose_email(customers_name, positions_source, positions, base_path):
    email_template_path = os.path.join(base_path, "email_template.html")

    email_template = ""
    with open(email_template_path, "r") as f:
        email_template = f.read()

    images_base_url = "https://ai-hue.ir/fenjan_phd_finder"

    logo_names = dir_list = os.listdir(os.path.join(base_path, "images/logo"))

    logo_path = f"{images_base_url}/logo/{random.choice(logo_names)}"
    email_template = email_template.replace("&heder_logo_place_holder", logo_path)

    greeting_image_names = dir_list = os.listdir(
        os.path.join(base_path, "images/greeting")
    )
    greeting_image_path = (
        f"{images_base_url}/greeting/{random.choice(greeting_image_names)}"
    )
    email_template = email_template.replace(
        "&greeting_image_place_holder", greeting_image_path
    )

    title_text = f"Ph.D. positions from {positions_source}"
    email_template = email_template.replace("&title_place_holder", title_text)

    today = datetime.today().strftime("%B %d, %Y")
    greeting_text = f"""Dear {customers_name},<br>
    I am pleased to present to you a list of Ph.D. positions that have been advertised on {positions_source} in the past 24 hours.<br><br>
    {today}"""
    email_template = email_template.replace("&greeting_place_holder", greeting_text)

    position_template = ""
    position_template_path = os.path.join(base_path, "position_template.html")
    with open(position_template_path, "r") as f:
        position_template = f.read()

    positions_template = ""
    _position_template = ""
    for position_num, position in enumerate(positions):
        _position_template = position_template.replace(
            "&position_title_place_holder", f"Ph.D. Position {position_num+1}"
        )
        position_summery = format_position_summery_text(position)
        _position_template = _position_template.replace(
            "&position_summery_place_holder", position_summery
        )
        positions_template += _position_template

    email_template = email_template.replace(
        "&position_template_place_holder", positions_template
    )

    footer_text = 'Developed by <a href="https://hue-salari.ir/" rel="noopener" style="text-decoration: none; color: #52a150;" target="_blank">Hue (MohammadHossein) Salari</a>'
    email_template = email_template.replace("&footer_place_holder", footer_text)
    return email_template


def compose_and_send_email(
    recipient_email, recipient_name, source_name, positions_list, utils_path
):
    """
    Compose and send email containing positions.

    Parameters:
        recipient_email (str): Email address to send the email to.
        recipient_name (str): Customer's name to include in the email.
        source_name (str): Name of the source of the positions
        positions_list (List[str]): List of positions to include in the email.
        utils_path (str): Base path for the email template file.

    Returns:
        None
    """
    # generate email text using the email template and the given positions

    emails_text = []
    for (title, url, matched_keywords, date), _ in positions_list:
        text = f'<span style="font-weight: bold">Title:</span> {title}\n'
        text += f'\n<span style="font-weight: bold">Matched Keywords:</span>\n<span style="color:#68677b">{matched_keywords}</span>'
        if date != "":
            text += f'\n<span style="font-weight: bold">Date:</span> {date}\n'
        text += f"<br><br>🔗: {url}"
        emails_text.append(text)
    email_text = compose_email(recipient_name, source_name, emails_text, utils_path)
    # send email with the generated text
    send_email(
        recipient_email,
        f"PhD Positions from {source_name}",
        email_text,
        "html",
    )


if __name__ == "__main__":

    import pickle

    with open("utils/test.positions", "rb") as f:
        positions = pickle.load(f)
    base_path = os.path.dirname(os.path.abspath(__file__))

    customers_name = "Hue Salari"
    positions_source = "Twitter"

    email_template = email_template = compose_email(
        customers_name, positions_source, positions[:20], base_path
    )
    out_path = "utils/test.html"
    print(f"[info]: html file saved in {out_path}")
    with open(out_path, "w") as f:
        f.write(email_template)
