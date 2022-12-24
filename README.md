# Fenjan

![](https://ai-hue.ir/fenjan_phd_finder/logo/logo06_r.png)

Fenjan is a Python bot that searches Twitter and LinkedIn for PhD positions and sends emails to customers based on their keywords.

## Installation

To install Fenjan, clone the repository and install the required dependencies:
```
git clone https://github.com/your-username/fenjan.git
cd fenjan
pip install -r requirements.txt
```
## Usage

To use Fenjan, you will need to obtain API keys for Twitter and LinkedIn. You can do this by creating a developer account on their respective websites.

Once you have your API keys, create a `.env` file in the root directory of the project and add your keys like so:

```

API_KEY = "your_twitter_api_key"
API_KEY_SECRET = "your_twitter_api_secret"
ACCESS_TOKEN = "your_twitter_access_token"
ACCESS_TOKEN_SECRET = "your_twitter_token_secret"


EMAIL_ADDRESS= "your_email_address"
EMAIL_PASSWORD= "your_email_api_password"


LINKEDIN_EMAIL_ADDRESS = "your_linkedin_email_address"
LINKEDIN_PASSWORD = "your_linkedin_password"
```

You can then run the bot using the following commands:

```
python linkedin.py
python twitter.py
```
