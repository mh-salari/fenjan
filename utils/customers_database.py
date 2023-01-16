import mysql.connector
import argparse
import json
import os

from dotenv import load_dotenv

from dataclasses import dataclass


@dataclass
class Customer:
    name: str
    email: str
    expiration_date: str
    keywords: list


class CustomerDatabase:
    def __init__(self, host, user, password, db, port):
        """
        Initializes the CustomerDatabase class

        Args:
            host (str): The hostname or IP address of the MySQL server.
            user (str): The username to connect to the MySQL server.
            password (str): The password to connect to the MySQL server.
            db (str): The name of the database.
        """
        self.host = host
        self.user = user
        self.password = password
        self.db = db
        self.port = port
        self.connection = None

    def connect_to_database(self):
        """
        Connects to the MySQL database and returns a connection object.

        Returns:
            connection : a connection object to the database
        """
        self.connection = mysql.connector.connect(
            host=self.host,
            user=self.user,
            password=self.password,
            database=self.db,
            port=self.port,
        )
        return self.connection

    def create_customers_table(self, table_name):
        """
        This function creates a table for customer data if it doesn't already exist.

        """
        # Create a cursor object to execute SQL commands
        cursor = self.connection.cursor()

        # Create the table if it doesn't already exist
        sql_query = f"""
        CREATE TABLE IF NOT EXISTS {table_name} (
            name VARCHAR(255),
            email VARCHAR(255) UNIQUE,
            expiration_date DATETIME,
            keywords JSON
        )
        """

        # Execute the SQL statement

        cursor.execute(sql_query)

        # Commit the changes to the database
        self.connection.commit()

        # Close the cursor
        cursor.close()

    def add_customer_data(self, table_name, data):
        """
        This function adds customer data to the specified table.

        Args:
            table_name (str): The name of the table to add data to.
            data (dict): A dictionary containing the customer data to add to the table.
        """
        # Create a cursor object to execute SQL commands
        cursor = self.connection.cursor()

        # Prepare the SQL statement to insert the data into the table
        sql_query = f"REPLACE INTO {table_name} (name, email, expiration_date, keywords) VALUES (%s, %s, %s, %s)"
        customer_data = (
            data.name,
            data.email,
            data.expiration_date,
            json.dumps(data.keywords),
        )

        # Execute the SQL statement
        cursor.execute(sql_query, customer_data)

        # Commit the changes to the database
        self.connection.commit()

        # Close the cursor
        cursor.close()

    def delete_customer_data(self, table_name, email):
        """
        This function deletes customer data from the specified table based on email.

        Args:
            table_name (str): The name of the table to delete data from.
            email (str): The email of the customer to delete.
        """
        # Create a cursor object to execute SQL commands
        cursor = self.connection.cursor()

        # Prepare the SQL statement to delete the data from the table
        sql_query = f"DELETE FROM {table_name} WHERE email = %s"
        customer_data = (email,)

        # Execute the SQL statement
        cursor.execute(sql_query, customer_data)

        # Commit the changes to the database
        self.connection.commit()

        # Close the cursor
        cursor.close()

    def get_customer_data(self, table_name):
        """
        This function connects to the MySQL database and retrieves the values from the 'positions' table.
        It returns a list of Position objects, where each object represents a row of the table.
        """
        cursor = self.connection.cursor()

        # SELECT statement to retrieve the values from the positions table
        query = f"SELECT * FROM {table_name}"
        cursor.execute(query)

        # Fetch the rows and create a list of Position objects
        customers = []
        for (name, email, expiration_date, keywords) in cursor:
            customers.append(
                Customer(name, email, expiration_date, json.loads(keywords))
            )

        # Close the cursor and connection
        cursor.close()

        return customers


def main():

    parser = argparse.ArgumentParser()
    parser.add_argument("--dotenv_path", help="path to .env file", required=True)
    parser.add_argument(
        "--action", help="action to perform: add, delete", choices=["add", "delete"]
    )
    parser.add_argument("--data", help="customer data in json format", type=json.loads)
    args = parser.parse_args()

    dotenv_path = args.dotenv_path
    load_dotenv(dotenv_path)
    # Get the database connection details from the .env file
    db_user = os.getenv("DB_USERNAME")
    db_password = os.getenv("DB_PASSWORD")
    db_host = os.getenv("DB_HOST")
    db_port = os.getenv("DB_PORT")
    db_name = os.getenv("DB_NAME")
    db_table = "customers"
    print(f"Connecting to {db_name}")
    db = CustomerDatabase(
        host=db_host, user=db_user, password=db_password, db=db_name, port=db_port
    )
    cnx = db.connect_to_database()
    print(f"Creating '{db_table}' table if not exist")
    db.create_customers_table(db_table)
    if args.action == "add":
        data = Customer(
            args.data["name"],
            args.data["email"],
            args.data["expiration_date"],
            args.data["keywords"],
        )
        print(f"Adding {data} to the database")
        db.add_customer_data(db_table, data)
    elif args.action == "delete":
        print(
            f"Deleting data of customer with email address: {args.data['email']} from the database"
        )
        db.delete_customer_data(db_table, args.data["email"])
    cnx.close()


if __name__ == "__main__":
    main()
#  python customers_database.py --dotenv_path "/media/hue/Data/codes/fenjan/.env"   --action add --data '{"name": "Hue Salari", "email": "hue.salari@gmail.com", "expiration_date": "2024-01-01", "keywords": ["computer vision","deep learning","machine learning","artificial intelligence","remote sensing","object detection","classification","segmentation","object extraction","supervised learning","semi-supervised learning","unsupervised learning","weakly-supervised learning","#AI","#ML"]}'
#  python customers_database.py --dotenv_path "/media/hue/Data/codes/fenjan/.env"   --action delete --data '{"email": "hue.salari@gmail.com"}'
