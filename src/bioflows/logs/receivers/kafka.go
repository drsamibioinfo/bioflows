package receivers

type KafkaReceiver struct {
	Config map[string]interface{}
}
func (k *KafkaReceiver) SetConfig(config map[string]interface{}) {
	k.Config = config
}

func (k *KafkaReceiver) Write(p []byte) (int , error) {
	return 0 , nil
}
