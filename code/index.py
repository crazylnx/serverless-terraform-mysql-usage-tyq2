# -*- coding: utf-8 -*-
import json
import logging
import pymysql
import os

logger = logging.getLogger()


def getConnection(config):
    try:
        conn = pymysql.connect(
            host=config["DB_HOST"]["value"],  # 替换为您的HOST名称。
            port=int(config["DB_PORT"]["value"]),  # 替换为您的端口号。
            user=config["DB_USER"]["value"],  # 替换为您的用户名。
            passwd=config["DB_PASSWORD"]["value"],  # 替换为您的用户名对应的密码。
            db=config["DATABASE_NAME"]["value"],  # 替换为您的数据库名称。
            connect_timeout=5)
        return conn
    except Exception as e:
        logger.error(e)
        logger.error(
            "ERROR: Unexpected error: Could not connect to MySql instance.")
        raise Exception(str(e))


def handler(event, context):
    output = os.environ['OUTPUT']
    config = json.loads(output)

    conn = getConnection(config)
    try:
        with conn.cursor() as cursor:
            cursor.execute("CREATE TABLE IF NOT EXISTS users (name VARCHAR(255), address VARCHAR(255))")
            cursor.execute("INSERT INTO users (name, address) VALUES (\"zhangsanfeng\", \"central block\")")
            conn.commit()
    except Exception as e:
        logger.error(e)
    finally:
        conn.close()
    #   check result
    conn = getConnection(config)
    res =''
    try:
        with conn.cursor() as cursor:
            cursor.execute('select * from users')
            res = cursor.fetchall()
            logger.info('Get result: ' + f"json: {json.dumps(res)}")
    except Exception as e:
        logger.error(e)
    finally:
        conn.close()
    return 'Get result: ' + f"json: {json.dumps(res)}"





