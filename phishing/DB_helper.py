import psycopg2
from configparser import ConfigParser

class DB:
    def __init__(self):   
        # read connection parameters
        params = self.config()
        # connect to the PostgreSQL server
        print('Connecting to the PostgreSQL database...')
        self.conn = psycopg2.connect(**params)
        # create a cursor
        self.cur = self.conn.cursor()
        # execute a statement to display the PostgreSQL database server version
        print('PostgreSQL database version:')
        self.cur.execute('SELECT version()')  
        print(self.cur.fetchone())

                       
    def config(self, filename='database.ini', section='postgresql'):
        # create a parser
        parser = ConfigParser()
        # read config file
        parser.read(filename)
        # get section, default to postgresql
        db = {}
        if parser.has_section(section):
            params = parser.items(section)
            for param in params:
                db[param[0]] = param[1]
        else:
            raise Exception('Section {0} not found in the {1} file'.format(section, filename))
        return db
    


    def close(self):
        self.cur.close()
        if self.conn is not None:
                self.conn.close()
                print('Database connection closed.')