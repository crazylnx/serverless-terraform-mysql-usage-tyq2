package terraform

func Apply(client *Client) (string, error) {
	err := client.init()
	if err != nil {
		return "", err
	}

	err, outputs := client.apply()
	if err != nil {
		return "", err
	}

	return outputs, nil
}

func Destroy(t *Client) (string, error) {
	if err := t.init(); err != nil {
		return "", err
	}

	err, outputs := t.destroy()
	if err != nil {
		return "", err
	}

	return outputs, nil
}
