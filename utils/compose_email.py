#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Des 6 2022
@author: Hue (MohammadHossein) Salari
@email:hue.salari@gmail.com
"""

from urlextract import URLExtract
from datetime import datetime
import random
import os


def format_position_summery_text(text):
    extractor = URLExtract()
    formatted_text = ""
    for line in text.splitlines(True):
        for word in line.split():
            url = extractor.find_urls(word)
            if word[0] == "#":
                formatted_text += (
                    '<span style="color:#68677b">' + word + "</span>" + " "
                )
            elif url:
                formatted_text += (
                    f'<a href="{url[0]}" rel="noopener" style="text-decoration: underline; color: #8a3b8f;" target="_blank">{url[0]}</a>'
                    + word.replace(url[0], "")
                )
            else:
                formatted_text += word + " "
        formatted_text += "<br>"

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

    title_text = "Ph.D. positions from Twitter"
    email_template = email_template.replace("&title_place_holder", title_text)

    today = datetime.today().strftime("%B %d, %Y")
    greeting_text = f"""Dear {customers_name},<br> 
    I am pleased to present to you a list of fully funded Ph.D. positions that have been advertised on {positions_source} in the past 24 hours.<br>
    {today}"""
    email_template = email_template.replace("&greeting_place_holder", greeting_text)

    position_template = ""
    position_template_path = "/media/hue/Data/codes/fenjan/utils/position_template.html"
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


if __name__ == "__main__":

    positions = [
        "date:2022-12-05 12:16:58\nby:CDT Artificial Intelligence+Music\nBe sure also to check out the list of potential PhD topics: https://t.co/9RcDLP0AUh\nhttps://twitter.com/twitter/statuses/1599739607284736001",
        "date:2022-12-05 07:34:49\nby:Onderzoeker\n#vacature PhD position: Artificial intelligence in stroke imaging. https://t.co/tLcvDWm4X4\nhttps://twitter.com/twitter/statuses/1599668599416922113",
        "date:2022-12-06 15:31:59\nby:Chirag Nagpal\n@Maggiemakar @UMichCSE Must mention Mingzhu Liu @MingzhuLiu4 (senior CS Major at UMich CS) is looking for a PhD position in ML for Health .. Mingzhu worked with us this summer and did fantastic work in using Deep Learning for Risk Estimation with Medical Imaging!\nhttps://twitter.com/twitter/statuses/1600151072974073856",
        "date:2022-12-06 11:51:13\nby:NZZ Jobs\nPhD Position: Project on Artificial Intelligence and People Analytics https://t.co/TsWo2S1bqM\nhttps://twitter.com/twitter/statuses/1600095512203759616",
        "date:2022-12-06 15:17:55\nby:Harquail School of Earth Sciences\nInterested in #DataAnalytics and #Geoscience? We have a @CFREF_APOGEE PhD opportunity starting in winter 2023 with @LUEngrCompSci - check out the posting &amp; apply now! https://t.co/ojJVWzBQJ3\n#MachineLearning #geospatial #randomforest #mapping #MetalEarth #geology #computerscience\nhttps://twitter.com/twitter/statuses/1600147531970715654",
        "date:2022-12-05 21:28:24\nby:Kate Allen @ Equibreathe\n\u2066@VetsSGD\u2069 PhD opportunity using machine learning approaches for analysis of equine endoscopy recordings. Bristol Veterinary School.  https://t.co/p7BtzsAY9D\nhttps://twitter.com/twitter/statuses/1599878379716632584",
    ]
    base_path = os.path.dirname(os.path.abspath(__file__))

    customers_name = "Hue Salari"
    positions_source = "Twitter"

    email_template = email_template = compose_email(
        customers_name, positions_source, positions, base_path
    )
    with open("utils/test.html", "w") as f:
        f.write(email_template)
