<?php
//BoltCall will send an api call to the bolt server
function BoltCall($boltserver_hostname, $boltserver_port, $apicall, $hmacname, $hmackey, $protocol, $message){

  // Create a json object with the message to encode and a timestamp
  date_default_timezone_set('UTC');
  $date = new DateTime();

  $payload = json_encode(array(
    "timestamp" => $date->format('U'),
    "message" => $message
  ));

  // Create the signed string
  $signature = hash_hmac('sha512', $payload, $hmackey, false);

  // Encode the payload and signature to Base64
  $basePayload = base64_encode($payload);
  $baseSignature = base64_encode($signature);

  // Combine the encoded message with the key signature
  $jsonStr = json_encode(array(
    'data' => $basePayload,
    'signature' => $baseSignature
  ));

  // Post to the boltengine api
  $url = $protocol."://".$boltserver_hostname.":".$boltserver_port.$apicall;
  echo "URL=".$url;
  $cred = sprintf('Authorization: Basic %s', base64_encode("$hmacname:password_ignored"));
  $opts = array(
    'http'=>array(
      'method'=>'POST',
      'header'=>$cred."\r\nContent-type: application/x-www-form-urlencoded"
                     ."\r\nContent-Length: ".strlen($jsonStr)."\r\n",
      'content'=>$jsonStr
    )
  );
  $ctx = stream_context_create($opts);
  $handle = fopen($url, 'r', false, $ctx);
  echo $payload;
  $stream = json_decode(stream_get_contents($handle));
  echo "\nFull response = ".json_encode($stream);

  $error = (array)$stream->error; // Cast to an array to determine if empty {}.
  if (!empty($error)) {
      echo "\nError: ".json_encode($stream->error)."\n";
      exit;
  }

  echo "\nreturn_value = ".json_encode($stream->return_value);
  // Iterate through the returned values
  /*
  foreach($stream->return_value as $key => $value) {
      echo "\nreturn_value->".$key."->custNumber = ".$value->custNumber;
      echo "\nreturn_value->".$key."->orderNumber = ".$value->orderNumber;
      echo "\nreturn_value->".$key."->gvcDate = ".$value->gvcDate;
      echo "\nreturn_value->".$key."->gvcFlag = ".$value->gvcFlag;
      echo "\nreturn_value->".$key."->cert1_Amount = ".$value->cert1_Amount;
      echo "\nreturn_value->".$key."->cert2_Amount = ".$value->cert2_Amount;
      echo "\nreturn_value->".$key."->cert3_Amount = ".$value->cert3_Amount;
      echo "\nreturn_value->".$key."->repc1_Amount = ".$value->repc1_Amount;
      echo "\nreturn_value->".$key."->repc2_Amount = ".$value->repc2_Amount;
      echo "\nreturn_value->".$key."->repc3_Amount = ".$value->repc3_Amount;
  }
  */
  echo "\n";
}
?>
