def search_for_keywords(text, keywords):
    """
    This function takes in a text and a list of keywords and searches for the keywords in the text.
    It returns a list of keywords that were found in the text.
    The search is case-insensitive.
    :param text: A string representing the text to search in.
    :param keywords: A list of keywords to search for.
    :return: A list of keywords found in the text.
    """
    results = [keyword.lower() in text.lower() for keyword in keywords]

    return [keywords[i] for i, x in enumerate(results) if x]
