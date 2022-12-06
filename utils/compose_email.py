from urlextract import URLExtract


def compose_email(positions):
    extractor = URLExtract()

    email_text = """
    <html>

    <head>
        <style>
        p {
            color: #03151e;
        }
        </style>
    </head>
    <body>
    """
    for num, position in enumerate(positions):
        email_text += "<section>"
        email_text += '<p style="white-space: pre-line;">'
        email_text += f'<span style="background-color:#9ac50d">Ph.D. Position {num+1} 🐱: \n</span>'
        for line in position.splitlines(True):
            for word in line.split():
                url = extractor.find_urls(word)
                if word[0] == "#":
                    email_text += (
                        '<span style="color:#2e6c87">' + word + "</span>" + " "
                    )
                elif url:
                    email_text += f"<a href={url[0]}>{url[0]}</a>" + word.replace(
                        url[0], ""
                    )
                else:
                    email_text += word + " "
            email_text += "\n"
        email_text += "</p>"
        email_text += "</section>"
    email_text += "</body>"

    return email_text
